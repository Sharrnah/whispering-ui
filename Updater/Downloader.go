package Updater

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

const defaultChunkSize int64 = 10 * 1024 * 1024 // 10 MB
const defaultConcurrentDownloads = 1

type OnProgress func(bytesWritten, contentLength uint64)

type WriteCounter struct {
	Total         uint64
	ContentLength uint64
	OnProgress    OnProgress
}

type Download struct {
	Url                 string
	FallbackUrls        []string
	Filepath            string
	ConcurrentDownloads int
	ChunkSize           int64 // in bytes
	WriteCounter        WriteCounter
	isResumed           bool
	serverResumeSupport bool
	maxRetries          int
	urlIndex            int
	mu                  sync.Mutex
	cond                *sync.Cond
	downloaded          map[int64][]byte
	nextWrite           int64
}

func (wc *WriteCounter) addBytes(n uint64) {
	wc.Total += n
	wc.OnProgress(wc.Total, wc.ContentLength)
}

func (d *Download) DownloadFile(retries int) error {
	d.downloaded = make(map[int64][]byte)
	d.maxRetries = retries
	if d.ConcurrentDownloads == 0 {
		d.ConcurrentDownloads = defaultConcurrentDownloads
	}
	if d.ChunkSize == 0 {
		d.ChunkSize = defaultChunkSize
	}
	return d.downloadFileWithRetry(retries)
}

func (d *Download) getCurrentUrl() string {
	currentUrl := d.Url
	if d.urlIndex > 0 && d.urlIndex <= len(d.FallbackUrls) {
		currentUrl = d.FallbackUrls[d.urlIndex-1]
	}
	return currentUrl
}

func (d *Download) retryAction(retries int, err error) error {
	currentUrl := d.getCurrentUrl()

	if retries > 0 {
		fmt.Printf("Error downloading %s: %s. Retrying in 1 seconds...\n", d.Url, err.Error())
		time.Sleep(2 * time.Second)
		return d.downloadFileWithRetry(retries - 1)
	} else {
		if d.urlIndex < len(d.FallbackUrls) {
			fmt.Printf("All retries for URL %s have failed. Trying the next fallback URL...\n", currentUrl)
			d.urlIndex++
			return d.downloadFileWithRetry(d.maxRetries)
		} else {
			fmt.Printf("All retries for URL %s and all fallback URLs have failed.\n", currentUrl)
			return err
		}
	}
}

type Chunk struct {
	data   []byte
	offset int64
}

func (d *Download) IsResuming() bool {
	return d.isResumed
}

func (d *Download) downloadFileWithRetry(retries int) error {
	currentUrl := d.getCurrentUrl()
	//d.downloaded = make([]int64, 0)

	req, err := http.NewRequest("HEAD", currentUrl, nil)
	if err != nil {
		return d.retryAction(retries, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return d.retryAction(retries, err)
	}
	defer resp.Body.Close()

	contentLength := resp.ContentLength
	d.WriteCounter.ContentLength = uint64(contentLength)

	// Define totalSize variable
	totalSize := int64(d.WriteCounter.ContentLength)

	// Check if the file already exists and get its size
	var startBytes int64 = 0
	if _, err := os.Stat(d.Filepath + ".tmp"); err == nil {
		startBytes = d.getFileSize(d.Filepath + ".tmp")
	}

	// Set ResumeSupport to true if the file download is resumed and the server supports resuming
	d.isResumed = startBytes > 0 && d.serverResumeSupport

	// Create the file without overwriting it
	out, err := os.OpenFile(d.Filepath+".tmp", os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer out.Close()

	//chunksCount := (contentLength + chunkSize - 1) / chunkSize

	// Create channels for communication
	chunksChannel := make(chan Chunk, d.ConcurrentDownloads)
	errorsChannel := make(chan error, d.ConcurrentDownloads)

	var wg sync.WaitGroup

	totalChunks := int(math.Ceil(float64(contentLength) / float64(d.ChunkSize)))
	startingChunk := startBytes / d.ChunkSize
	remainingChunks := int64(totalChunks - int(startBytes/d.ChunkSize))

	// Initialize the WriteCounter values
	d.WriteCounter.Total = uint64(startBytes)
	d.WriteCounter.ContentLength = uint64(totalSize)

	// Initialize d.nextWrite
	d.nextWrite = startBytes

	// Concurrent download loop
	for i := 0; i < d.ConcurrentDownloads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				chunkIndex := atomic.AddInt64(&startingChunk, 1) - 1
				if chunkIndex >= int64(totalChunks) {
					break
				}

				remaining := atomic.AddInt64(&remainingChunks, -1)
				if remaining < 0 {
					break
				}

				start := chunkIndex * d.ChunkSize // Updated start calculation
				end := start + d.ChunkSize - 1
				if end >= totalSize {
					end = totalSize - 1
				}

				chunk, downloaded, err := d.downloadChunk(currentUrl, start, end)
				if err != nil {
					errorsChannel <- err
					return
				}

				if downloaded {
					chunksChannel <- *chunk
				}
			}
		}()
	}

loop:
	for {
		select {
		case err := <-errorsChannel:
			return d.retryAction(retries, err)
		case chunk := <-chunksChannel:
			d.mu.Lock()
			d.downloaded[chunk.offset] = chunk.data

			for {
				data, ok := d.downloaded[d.nextWrite]
				if !ok {
					break
				}

				_, err := out.Write(data)
				if err != nil {
					d.mu.Unlock()
					return err
				}

				d.WriteCounter.addBytes(uint64(len(data)))
				delete(d.downloaded, d.nextWrite)
				d.nextWrite += int64(len(data))
			}

			if d.nextWrite == totalSize {
				d.mu.Unlock()
				break loop
			}

			d.mu.Unlock()
		}
	}

	wg.Wait()

	// Close the file without defer so it can happen before Rename()
	out.Close()

	if err = os.Rename(d.Filepath+".tmp", d.Filepath); err != nil {
		return err
	}
	return nil
}

func (d *Download) downloadChunk(url string, start, end int64) (*Chunk, bool, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusPartialContent {
		d.serverResumeSupport = true
	} else if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, false, err
	}

	return &Chunk{
		offset: start,
		data:   data,
	}, true, nil
}

func (d *Download) getFileSize(filepath string) int64 {
	file, err := os.Open(filepath)
	if err != nil {
		return 0
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return 0
	}
	return info.Size()
}
