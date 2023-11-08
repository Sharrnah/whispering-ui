package ModelDownloader

type modelLink struct {
	urls     []string `yaml:"urls"`
	checksum string   `yaml:"checksum"`
}
type modelNameLinks struct {
	cachePath string                `yaml:"cachePath"`
	modelLink map[string]*modelLink `yaml:"modelLink"`
}

type modelNameLinksMap map[string]*modelNameLinks

var modelNameLinksList = modelNameLinksMap{
	"Whisper": {
		cachePath: "whisper",
		modelLink: map[string]*modelLink{
			"tiny.en": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/d3dd57d32accea0b295c96e26691aa14d8822fac7d9d27d5dc00b4ca2826dd03/tiny.en.pt",
				},
				checksum: "d3dd57d32accea0b295c96e26691aa14d8822fac7d9d27d5dc00b4ca2826dd03",
			},
			"tiny": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/65147644a518d12f04e32d6f3b26facc3f8dd46e5390956a9424a650c0ce22b9/tiny.pt",
				},
				checksum: "65147644a518d12f04e32d6f3b26facc3f8dd46e5390956a9424a650c0ce22b9",
			},
			"base.en": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/25a8566e1d0c1e2231d1c762132cd20e0f96a85d16145c3a00adf5d1ac670ead/base.en.pt",
				},
				checksum: "25a8566e1d0c1e2231d1c762132cd20e0f96a85d16145c3a00adf5d1ac670ead",
			},
			"base": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/ed3a0b6b1c0edf879ad9b11b1af5a0e6ab5db9205f891f668f8b0e6c6326e34e/base.pt",
				},
				checksum: "ed3a0b6b1c0edf879ad9b11b1af5a0e6ab5db9205f891f668f8b0e6c6326e34e",
			},
			"small.en": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/f953ad0fd29cacd07d5a9eda5624af0f6bcf2258be67c92b79389873d91e0872/small.en.pt",
				},
				checksum: "f953ad0fd29cacd07d5a9eda5624af0f6bcf2258be67c92b79389873d91e0872",
			},
			"small": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/9ecf779972d90ba49c06d968637d720dd632c55bbf19d441fb42bf17a411e794/small.pt",
				},
				checksum: "9ecf779972d90ba49c06d968637d720dd632c55bbf19d441fb42bf17a411e794",
			},
			"medium.en": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/d7440d1dc186f76616474e0ff0b3b6b879abc9d1a4926b7adfa41db2d497ab4f/medium.en.pt",
				},
				checksum: "d7440d1dc186f76616474e0ff0b3b6b879abc9d1a4926b7adfa41db2d497ab4f",
			},
			"medium": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/345ae4da62f9b3d59415adc60127b97c714f32e89e936602e85993674d08dcb1/medium.pt",
				},
				checksum: "345ae4da62f9b3d59415adc60127b97c714f32e89e936602e85993674d08dcb1",
			},
			"large-v1": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/e4b87e7e0bf463eb8e6956e646f1e277e901512310def2c24bf0e11bd3c28e9a/large-v1.pt",
				},
				checksum: "e4b87e7e0bf463eb8e6956e646f1e277e901512310def2c24bf0e11bd3c28e9a",
			},
			"large-v2": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/81f7c96c852ee8fc832187b0132e569d6c3065a3252ed18e56effd0b6a73e524/large-v2.pt",
				},
				checksum: "81f7c96c852ee8fc832187b0132e569d6c3065a3252ed18e56effd0b6a73e524",
			},
			"large-v3": {
				urls: []string{
					"https://openaipublic.azureedge.net/main/whisper/models/e5b1a55b89c1367dacf97e3e19bfd829a01529dbfdeefa8caeb59b3f1b81dadb/large-v3.pt",
				},
				checksum: "e5b1a55b89c1367dacf97e3e19bfd829a01529dbfdeefa8caeb59b3f1b81dadb",
			},
		},
	},
	"WhisperCT2": {
		cachePath: "whisper",
		modelLink: map[string]*modelLink{
			"tiny_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tiny-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tiny-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tiny-ct2-fp16.zip",
				},
				checksum: "3c7c0512b7b881ecb4cb0693d543aed2a9178968bef255fa0ca8b880541ec789",
			},
			"tiny_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tiny-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tiny-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tiny-ct2.zip",
				},
				checksum: "18f4d5a6dbb9d27b748ee7a58ef455ff6640f230e5d64781e9cfb16181136b04",
			},
			"tiny.en_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tiny.en-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tiny.en-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tiny.en-ct2-fp16.zip",
				},
				checksum: "a14fedc8e57090505ec46119d346895604f5a6b5a8a44a7a137c44169544ea99",
			},
			"tiny.en_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tiny.en-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tiny.en-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tiny.en-ct2.zip",
				},
				checksum: "814c670c9922574c9e0e3be8d7f616e53347ec2dee099648523e2f88ec436eec",
			},
			"base_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/base-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/base-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/base-ct2-fp16.zip",
				},
				checksum: "fa863d01b4ef07bab0467d13b33221c8e6273362078ec6268bbc6398f40c0ab4",
			},
			"base_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/base-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/base-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/base-ct2.zip",
				},
				checksum: "e95001e10c40b57797e208f2e915e16d86bac67f204742bac2b8950e6eeb3539",
			},
			"base.en_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/base.en-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/base.en-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/base.en-ct2-fp16.zip",
				},
				checksum: "ec00c31ef78f035950c276ff01e5da96b4e9761bc15e872b2ec02371ac357484",
			},
			"base.en_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/base.en-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/base.en-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/base.en-ct2.zip",
				},
				checksum: "5113b44b8f4fe1927f935d85326df5bbe708ab269144fc9399234f9e9b9d61d1",
			},
			"small_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/small-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/small-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/small-ct2-fp16.zip",
				},
				checksum: "9f0618523bf19dc68d99109ba319f2faba2c94ef9d063aa300115935f3d09f14",
			},
			"small_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/small-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/small-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/small-ct2.zip",
				},
				checksum: "b887054992cf42abddad057e4b52f3ef6b1a079485244d786f1941a6fec8c02e",
			},
			"small.en_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/small.en-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/small.en-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/small.en-ct2-fp16.zip",
				},
				checksum: "9f0618523bf19dc68d99109ba319f2faba2c94ef9d063aa300115935f3d09f14",
			},
			"small.en_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/small.en-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/small.en-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/small.en-ct2.zip",
				},
				checksum: "c7eeb56070467bfad17ec774f66ce8dfc0b601d9c2ad5f96b3e4da9331552692",
			},
			"medium_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/medium-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/medium-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/medium-ct2-fp16.zip",
				},
				checksum: "13d2d91bdd2c3722c0592cbffca468992257eb3ddb782b1779c59091a4d91dd4",
			},
			"medium_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/medium-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/medium-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/medium-ct2.zip",
				},
				checksum: "5682a3833f4c87ed749778a844ccc9da6d8b3e3a2fef338cf5e66b495050e2e6",
			},
			"medium.en_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/medium.en-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/medium.en-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/medium.en-ct2-fp16.zip",
				},
				checksum: "13d2d91bdd2c3722c0592cbffca468992257eb3ddb782b1779c59091a4d91dd4",
			},
			"medium.en_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/medium.en-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/medium.en-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/medium.en-ct2.zip",
				},
				checksum: "8bf93eb5018c44c9115b6b942f8bc518790f88c2db93920f2da1a6a1efefe002",
			},
			"large-v1_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/large-v1-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/large-v1-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/large-v1-ct2-fp16.zip",
				},
				checksum: "42ecc70522602e69fe6365ef73173bbb1178ff8fd99210b96ea9025a205014bb",
			},
			"large-v1_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/large-v1-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/large-v1-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/large-v1-ct2.zip",
				},
				checksum: "82bd59ee73d7b52f60de5566e8e3e429374bd2dd1bce3e2f6fc18b620dbcf0cf",
			},
			"large-v2_float16": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/large-v2-ct2-fp16.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/large-v2-ct2-fp16.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/large-v2-ct2-fp16.zip",
				},
				checksum: "2397ed6433a08d4b6968852bc1b761b488c3149a3a52f49b62b2ac60d1d5cef0",
			},
			"large-v2_float32": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/large-v2-ct2.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/large-v2-ct2.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/large-v2-ct2.zip",
				},
				checksum: "c9e889f59cacfef9ebe76a1db5d80befdcf0043195c07734f6984d19e78c8253",
			},
		},
	},
	"WhisperCT2_Tokenizer": {
		cachePath: "whisper",
		modelLink: map[string]*modelLink{
			"normal": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tokenizer.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tokenizer.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tokenizer.zip",
				},
				checksum: "f6233d181a04abce6e2ba20189d5872b58ce2e14917af525a99feb5619777d7d",
			},
			"en": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/Whisper-CT2/tokenizer.en.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/Whisper-CT2/tokenizer.en.zip",
					"https://s3.libs.space:9000/ai-models/Whisper-CT2/tokenizer.en.zip",
				},
				checksum: "fb364e7cae84eedfd742ad116a397daa75e4eebba38f27e3f391ae4fee19afa9",
			},
		},
	},
	"NLLB200CT2": {
		cachePath: "nllb200_ct2",
		modelLink: map[string]*modelLink{
			"small": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/NLLB-200/CT2/small.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/NLLB-200/CT2/small.zip",
					"https://s3.libs.space:9000/ai-models/NLLB-200/CT2/small.zip",
				},
				checksum: "54188e59e5267329996f93a559befc0c14c09ef6a4f5f4e9645b0da94e380d47",
			},
			"medium": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/NLLB-200/CT2/medium.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/NLLB-200/CT2/medium.zip",
					"https://s3.libs.space:9000/ai-models/NLLB-200/CT2/medium.zip",
				},
				checksum: "88efd459f37d098bc44262721add08c57d22e482aab986edb4c7cbde5bd17cf9",
			},
			"large": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/NLLB-200/CT2/large.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/NLLB-200/CT2/large.zip",
					"https://s3.libs.space:9000/ai-models/NLLB-200/CT2/large.zip",
				},
				checksum: "c1f5618552cdfad2a5daf74e8218e5c583a6ee10acd3b8dc139ae2d94067af85",
			},
		},
	},
	"sentencepiece": {
		modelLink: map[string]*modelLink{
			"default": {
				urls: []string{
					"https://usc1.contabostorage.com/8fcf133c506f4e688c7ab9ad537b5c18:ai-models/NLLB-200/CT2/sentencepiece.zip",
					"https://eu2.contabostorage.com/bf1a89517e2643359087e5d8219c0c67:ai-models/NLLB-200/CT2/sentencepiece.zip",
					"https://s3.libs.space:9000/ai-models/NLLB-200/CT2/sentencepiece.zip",
				},
				checksum: "7e7fe41261d253ebba549de48b280021b1ae9d7915aa583689b34aa1f8604d13",
			},
		},
	},
}
