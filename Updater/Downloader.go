package Updater

import (
	"context"
	"fmt"
	"fyne.io/fyne/v2"
	"github.com/hashicorp/go-cleanhttp"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"whispering-tiger-ui/Utilities"
)

const DefaultChunkSize int64 = 20 * 1024 * 1024 // 20 MB
const defaultConcurrentDownloads = 1

var netClient = cleanhttp.DefaultPooledClient()

type OnProgress func(bytesWritten, contentLength uint64, speed float64)

type WriteCounter struct {
	Total         uint64
	ContentLength uint64
	OnProgress    OnProgress
	startTime     time.Time
	LastUpdate    time.Time
	speedMA       *MovingAverage
}

type Download struct {
	skipServerChan         chan bool
	cancelFunc             context.CancelFunc
	progressCtx            context.Context
	Url                    string
	FallbackUrls           []string
	UseMultiServerDownload bool
	Filepath               string
	ConcurrentDownloads    int
	ChunkSize              int64 // in bytes
	WriteCounter           WriteCounter
	isResumed              bool
	serverResumeSupport    bool
	maxRetries             int
	urlIndex               int
	mu                     sync.Mutex
	cond                   *sync.Cond
	downloaded             map[int64][]byte
	nextWrite              int64
	remoteFileSize         int64
}

func (d *Download) getUserAgent() string {
	// convert build int to string
	build := fyne.CurrentApp().Metadata().Build
	buildStr := fmt.Sprintf("%d", build)

	return "Whispering_Tiger_DL/" + fyne.CurrentApp().Metadata().Version + " (" + buildStr + ")"
}

func (d *Download) getRemoteFileSize() (int64, error) {
	currentUrl := d.getCurrentUrl()

	req, err := http.NewRequest("HEAD", currentUrl, nil)
	if err != nil {
		return 0, err
	}
	req.Header.Set("User-Agent", d.getUserAgent())

	resp, err := netClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.ContentLength, nil
}

func (d *Download) getRemoteFileSizeWithRetry(retries int) (int64, error) {
	for i := 0; i <= retries; i++ {
		remoteFileSize, err := d.getRemoteFileSize()
		if err == nil {
			return remoteFileSize, nil
		}

		if i < retries {
			fmt.Printf("Error getting remote file size %s: %s. Retrying in 1 second...\n", d.Url, err.Error())
			time.Sleep(1 * time.Second)
		} else {
			// Switch to the next fallback url if available
			if d.urlIndex < len(d.FallbackUrls) {
				fmt.Printf("All retries for URL %s have failed. Trying the next fallback URL...\n", d.getCurrentUrl())
				d.urlIndex++
				i = -1 // reset retry count for the next url
				continue
			} else {
				fmt.Printf("All retries for URL %s and all fallback URLs have failed.\n", d.getCurrentUrl())
				return 0, err
			}
		}
	}
	return 0, fmt.Errorf("Failed to get remote file size after %d retries", retries)
}

func (d *Download) addBytes(n uint64) {
	if n > 0 {
		d.WriteCounter.Total += n
	}
	if time.Since(d.WriteCounter.LastUpdate).Seconds() >= 1 || n == 0 {
		elapsed := time.Since(d.WriteCounter.startTime).Seconds()
		speed := float64(d.WriteCounter.Total) / elapsed
		d.WriteCounter.speedMA.Add(speed)
		avgSpeed := d.WriteCounter.speedMA.Average()
		d.WriteCounter.OnProgress(d.WriteCounter.Total, d.WriteCounter.ContentLength, avgSpeed)
		d.WriteCounter.LastUpdate = time.Now()
	}
}

func (d *Download) GetTotalDownloadedSize() int64 {
	// Calculate total size of chunks in memory
	totalSizeInMemory := 0
	for _, data := range d.downloaded {
		totalSizeInMemory += len(data)
	}

	// Calculate the total downloaded size as the sum of the size of the temporary file and the total size of the chunks in memory
	return d.getFileSize(d.Filepath+".tmp") + int64(totalSizeInMemory)
}

func (d *Download) DownloadFile(retries int) error {
	progressCtx, progressCancel := context.WithCancel(context.Background())
	d.cancelFunc = progressCancel
	d.progressCtx = progressCtx
	defer progressCancel()

	go func() {
		for {
			select {
			case <-time.After(1 * time.Second):
				d.addBytes(0)
			case <-progressCtx.Done():
				return
			}
		}
	}()

	d.downloaded = make(map[int64][]byte)
	d.maxRetries = retries
	if d.ConcurrentDownloads == 0 {
		d.ConcurrentDownloads = defaultConcurrentDownloads
	}
	if d.ChunkSize == 0 {
		d.ChunkSize = DefaultChunkSize
	}
	d.WriteCounter.speedMA = NewMovingAverage(movingAverageWindow) // 10 is the moving average window size
	d.WriteCounter.startTime = time.Now()                          // Record the start time when download starts

	// check if the server file is smaller than the local file (which means something is wrong)
	// Call getRemoteFileSize to get the size of the remote file
	err := error(nil)
	d.remoteFileSize, err = d.getRemoteFileSizeWithRetry(retries)
	if err != nil {
		return err
	}

	// If remote file size is -1, proceed to download without knowing the file size
	if d.remoteFileSize == -1 {
		d.ChunkSize = math.MaxInt64                   // set the chunk size to maximum value
		d.WriteCounter.ContentLength = math.MaxUint64 // set the content length to maximum value
		if Utilities.FileExists(d.Filepath + ".tmp") {
			err := os.Remove(d.Filepath + ".tmp")
			if err != nil {
				return err
			}
		}
	} else {
		// Check if the local file exists and if it's larger than the remote file
		localFileSize := d.getFileSize(d.Filepath + ".tmp")
		if Utilities.FileExists(d.Filepath+".tmp") && localFileSize > d.remoteFileSize {
			// If the local file is larger than the remote file, delete the local file
			err := os.Remove(d.Filepath + ".tmp")
			if err != nil {
				return err
			}
		}
	}

	return d.downloadFileWithRetry(retries, progressCtx, progressCancel)
}

func (d *Download) getCurrentUrl() string {
	currentUrl := d.Url
	if d.urlIndex > 0 && d.urlIndex <= len(d.FallbackUrls) {
		currentUrl = d.FallbackUrls[d.urlIndex-1]
	}
	return currentUrl
}

func (d *Download) retryAction(retries int, err error, progressCtx context.Context, contextCancel context.CancelFunc) error {
	currentUrl := d.getCurrentUrl()

	if retries > 0 {
		fmt.Printf("Error downloading %s: %s. Retrying in 1 seconds...\n", d.Url, err.Error())
		time.Sleep(2 * time.Second)
		return d.downloadFileWithRetry(retries-1, progressCtx, contextCancel)
	} else {
		if d.urlIndex < len(d.FallbackUrls) {
			fmt.Printf("All retries for URL %s have failed. Trying the next fallback URL...\n", currentUrl)
			d.urlIndex++
			return d.downloadFileWithRetry(d.maxRetries, progressCtx, contextCancel)
		} else {
			fmt.Printf("All retries for URL %s and all fallback URLs have failed.\n", currentUrl)
			return err
		}
	}
}

func (d *Download) GetDownloadUrlSubdomain() string {
	// get subdomain from download url
	downloadUrl := d.getCurrentUrl()
	return downloadUrl[strings.Index(downloadUrl, "://")+3 : strings.Index(downloadUrl, ".")]
}

func (d *Download) switchToNextUrl() {
	d.urlIndex = (d.urlIndex + 1) % (len(d.FallbackUrls) + 1)
}

func (d *Download) RestartDownloadWithNextServer() error {
	if d.skipServerChan != nil {
		d.skipServerChan <- true
		time.Sleep(100 * time.Millisecond) // Allow goroutines to process the skip signal
	}
	if d.cancelFunc != nil {
		d.cancelFunc()
	}
	d.switchToNextUrl()

	// Create a new context for the download
	d.progressCtx, d.cancelFunc = context.WithCancel(context.Background())

	return d.downloadChunkedFile(d.maxRetries, d.progressCtx, d.cancelFunc)
}

type Chunk struct {
	data   []byte
	offset int64
}

func (d *Download) IsResuming() bool {
	return d.isResumed
}

func (d *Download) downloadFullFile(url string) error {
	out, err := os.Create(d.Filepath + ".tmp")
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", d.getUserAgent())
	resp, err := netClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (d *Download) downloadChunkedFile(retries int, progressCtx context.Context, contextCancel context.CancelFunc) error {
	allUrls := append([]string{d.Url}, d.FallbackUrls...)
	if !d.UseMultiServerDownload {
		allUrls = []string{d.getCurrentUrl()}
	}
	if len(d.FallbackUrls) > 0 && d.UseMultiServerDownload {
		rand.Shuffle(len(allUrls), func(i, j int) { allUrls[i], allUrls[j] = allUrls[j], allUrls[i] })
	}
	currentUrl := d.getCurrentUrl()

	contentLength := d.remoteFileSize

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
				select {
				case <-d.skipServerChan:
					return
				case <-progressCtx.Done():
					return
				default:
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

					if !d.UseMultiServerDownload {
						currentUrl = d.getCurrentUrl()
					} else {
						// cycle through the servers in allUrls in a round-robin fashion.
						currentUrl = allUrls[chunkIndex%int64(len(allUrls))]
					}
					println("Downloading chunk %d of %d from %s", chunkIndex, totalChunks, currentUrl)

					chunk, downloaded, err := d.downloadChunk(currentUrl, start, end)
					if err != nil {
						errorsChannel <- err
						return
					}

					if downloaded {
						chunksChannel <- *chunk
					}
				}
			}
		}()
	}

	receivedChunks := 0

loop:
	for {
		select {
		case <-d.skipServerChan:
			d.mu.Unlock()
			return nil
		case err := <-errorsChannel:
			return d.retryAction(retries, err, progressCtx, contextCancel)
		case chunk := <-chunksChannel:
			receivedChunks++ // Increment the counter for received chunks

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

				d.addBytes(uint64(len(data)))
				delete(d.downloaded, d.nextWrite)
				d.nextWrite += int64(len(data))
			}

			if d.nextWrite == totalSize {
				contextCancel() // cancel the progress update goroutine
				d.addBytes(0)   // force progress update when finished
				d.mu.Unlock()
				break loop
			}

			d.mu.Unlock()
		}
	}

	wg.Wait()

	// Close the file without defer so it can happen before Rename()
	if err := out.Close(); err != nil {
		return err
	}
	return nil
}

func (d *Download) downloadFileWithRetry(retries int, progressCtx context.Context, contextCancel context.CancelFunc) error {
	currentUrl := d.getCurrentUrl()

	err := error(nil)

	contentLength := d.remoteFileSize

	d.WriteCounter.ContentLength = uint64(contentLength)

	if contentLength == -1 {
		// If the server doesn't support resuming, download the full file
		println("Server doesn't support resuming. downloading full file...")
		err = d.downloadFullFile(currentUrl)
		if err != nil {
			return err
		}
	} else {
		err = d.downloadChunkedFile(retries, progressCtx, contextCancel)
		if err != nil {
			return err
		}
	}

	// Maximum number of retries for rename
	maxRetries := 5

	// Time to wait between rename retries
	retryWait := time.Second

	var renameErr error
	for i := 0; i < maxRetries; i++ {
		renameErr = os.Rename(d.Filepath+".tmp", d.Filepath)
		if renameErr == nil {
			break
		}
		// The error occurred, wait for a bit before trying again
		time.Sleep(retryWait)
	}
	if renameErr != nil {
		return renameErr
	}
	return nil
}

func (d *Download) downloadChunk(url string, start, end int64) (*Chunk, bool, error) {
	select {
	case <-d.skipServerChan:
		return nil, false, fmt.Errorf("download was cancelled")
	default:
		// Continue with the download
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	req.Header.Set("User-Agent", d.getUserAgent())

	resp, err := netClient.Do(req)
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

// ################
// moving average
// ################

const movingAverageWindow = 5

type MovingAverage struct {
	size  int
	sum   float64
	queue []float64
}

func NewMovingAverage(size int) *MovingAverage {
	return &MovingAverage{
		size:  size,
		sum:   0.0,
		queue: make([]float64, 0, size),
	}
}

func (ma *MovingAverage) Add(value float64) {
	if len(ma.queue) >= ma.size {
		ma.sum -= ma.queue[0]
		ma.queue = ma.queue[1:]
	}
	ma.queue = append(ma.queue, value)
	ma.sum += value
}

func (ma *MovingAverage) Average() float64 {
	if len(ma.queue) == 0 {
		return 0
	}
	return ma.sum / float64(len(ma.queue))
}
