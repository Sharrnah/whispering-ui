package ProfileForm

import (
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/Hardwareinfo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

// Coordinator centralizes dynamic behaviors: enabling/disabling, option fallback, and multi-modal sync.
type Coordinator struct {
	Controls             *AllProfileControls
	IsLoadingSettings    *bool
	ComputeCapability    float32
	CPUMemoryBar         *widget.ProgressBar
	GPUMemoryBar         *widget.ProgressBar
	TotalGPUMemoryMiB    int64
	InProgrammaticUpdate bool
	SuppressPrompts      bool
	MirrorLocked         map[string]bool
}

func (c *Coordinator) getParentWindow() fyne.Window {
	wins := fyne.CurrentApp().Driver().AllWindows()
	if len(wins) == 0 {
		return nil
	}
	return wins[len(wins)-1]
}

func (c *Coordinator) SetOptionsWithFallback(sel *CustomWidget.TextValueSelect, options []CustomWidget.TextValueOption) {
	if sel == nil {
		return
	}
	current := sel.GetSelected()
	sel.Options = options
	if current == nil || !sel.ContainsEntry(current, CustomWidget.CompareValue) {
		if len(options) > 0 {
			sel.SetSelectedIndex(0)
		}
	} else {
		sel.SetSelected(current.Value)
	}
	sel.Refresh()
}

func (c *Coordinator) Enable(widgets ...fyne.Disableable) {
	for _, w := range widgets {
		if w != nil {
			w.Enable()
		}
	}
}
func (c *Coordinator) Disable(widgets ...fyne.Disableable) {
	for _, w := range widgets {
		if w != nil {
			w.Disable()
		}
	}
}

func (c *Coordinator) IsMultiModalModelPair(select1, select2 *CustomWidget.TextValueSelect) bool {
	if select1.GetSelected() == nil || select2.GetSelected() == nil {
		return false
	}
	model1 := select1.GetSelected().Value
	model2 := select2.GetSelected().Value
	multiModalModels := []string{"seamless_m4t", "phi4", "voxtral"}
	for _, m := range multiModalModels {
		if model1 == m && model2 == m {
			return true
		}
	}
	return false
}

func (c *Coordinator) SyncMultiModalTargets(sourceValue string,
	sourcePrecision, sourceDevice *CustomWidget.TextValueSelect,
	targetSize, targetPrecision, targetDevice *CustomWidget.TextValueSelect,
) {
	if targetSize != nil {
		targetSize.SetSelected(sourceValue)
		if d, s := sourceDevice.GetSelected(), targetDevice; d != nil && s != nil && s.ContainsEntry(d, CustomWidget.CompareValue) {
			s.SetSelected(d.Value)
		}
	}
	if sourcePrecision != nil && targetPrecision != nil {
		if sp := sourcePrecision.GetSelected(); sp != nil && targetPrecision.ContainsEntry(sp, CustomWidget.CompareValue) {
			targetPrecision.SetSelected(sp.Value)
		}
	}
	c.Disable(targetSize, targetPrecision, targetDevice)
}

func (c *Coordinator) HandleMultiModalModelSync(
	modelSelect1, modelSelect2 *CustomWidget.TextValueSelect,
	sourceValue string,
	sourcePrecision, sourceDevice *CustomWidget.TextValueSelect,
	targetSize, targetPrecision, targetDevice *CustomWidget.TextValueSelect,
) {
	if c.IsMultiModalModelPair(modelSelect1, modelSelect2) {
		c.SyncMultiModalTargets(sourceValue, sourcePrecision, sourceDevice, targetSize, targetPrecision, targetDevice)
		return
	}
	if modelSelect1.GetSelected() != nil && modelSelect1.GetSelected().Value != "" && modelSelect2.GetSelected() != nil && modelSelect2.GetSelected().Value != "" {
		c.Enable(targetSize, targetPrecision, targetDevice)
	}
}

func (c *Coordinator) ConfirmOption(titleKey, messageKey string, onYes func()) {
	if c.IsLoadingSettings != nil && *c.IsLoadingSettings {
		return
	}
	parent := c.getParentWindow()
	dialog.NewConfirm(lang.L(titleKey), lang.L(messageKey), func(b bool) {
		if !b {
			return
		}
		fyne.Do(func() {
			prevProg, prevPrompts := c.InProgrammaticUpdate, c.SuppressPrompts
			c.InProgrammaticUpdate, c.SuppressPrompts = true, true
			onYes()
			c.HandleMultiModalAllSync()
			c.SuppressPrompts, c.InProgrammaticUpdate = prevPrompts, prevProg
		})
	}, parent).Show()
}

func (c *Coordinator) EnsurePrecisionDeviceCompatibility(device, precision string) {
	switch device {
	case "cpu":
		if precision == "float16" || precision == "int8_float16" {
			dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "float16"}), c.getParentWindow())
		}
		if precision == "bfloat16" || precision == "int8_bfloat16" {
			dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CPU's", "Precision": "bfloat16"}), c.getParentWindow())
		}
	case "cuda":
		if precision == "int16" {
			dialog.ShowInformation(lang.L("Information"), lang.L("Most Devices of this type do not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU's", "Precision": "int16"}), c.getParentWindow())
		}
		if (precision == "bfloat16" || precision == "int8_bfloat16") && c.ComputeCapability < 8.0 {
			dialog.ShowInformation(lang.L("Information"), lang.L("Your Device most likely does not support this precision computation. Please consider switching to some other precision.", map[string]interface{}{"Device": "CUDA GPU", "Precision": "bfloat16"}), c.getParentWindow())
		}
	}
}

var multiModalCapabilities = map[string]struct{ STT, TXT, OCR bool }{
	"seamless_m4t": {STT: true, TXT: true, OCR: false},
	"voxtral":      {STT: true, TXT: true, OCR: false},
	"phi4":         {STT: true, TXT: true, OCR: true},
}

const (
	groupSTT = "STT"
	groupTXT = "TXT"
	groupOCR = "OCR"
)

func (c *Coordinator) getSelectedType(sel *CustomWidget.TextValueSelect) string {
	if sel == nil || sel.GetSelected() == nil {
		return ""
	}
	return sel.GetSelected().Value
}
func (c *Coordinator) getGroupType(g string) string {
	switch g {
	case groupSTT:
		return c.getSelectedType(c.Controls.STTType)
	case groupTXT:
		return c.getSelectedType(c.Controls.TxtType)
	case groupOCR:
		return c.getSelectedType(c.Controls.OCRType)
	}
	return ""
}
func (c *Coordinator) enableGroup(g string) {
	switch g {
	case groupSTT:
		c.Enable(c.Controls.STTModelSize, c.Controls.STTPrecision, c.Controls.STTDevice)
	case groupTXT:
		c.Enable(c.Controls.TxtSize, c.Controls.TxtPrecision, c.Controls.TxtDevice)
	case groupOCR:
		c.Enable(c.Controls.OCRPrecision, c.Controls.OCRDevice)
	}
}
func (c *Coordinator) disableGroup(g string) {
	switch g {
	case groupSTT:
		c.Disable(c.Controls.STTModelSize, c.Controls.STTPrecision, c.Controls.STTDevice)
	case groupTXT:
		c.Disable(c.Controls.TxtSize, c.Controls.TxtPrecision, c.Controls.TxtDevice)
	case groupOCR:
		c.Disable(c.Controls.OCRPrecision, c.Controls.OCRDevice)
	}
}

func (c *Coordinator) mirrorFromTo(controller, target string) {
	if controller == groupSTT && target == groupTXT {
		if c.Controls.STTModelSize != nil && c.Controls.STTModelSize.GetSelected() != nil && c.Controls.TxtSize != nil {
			size := c.Controls.STTModelSize.GetSelected()
			if c.Controls.TxtSize.ContainsEntry(size, CustomWidget.CompareValue) {
				c.Controls.TxtSize.SetSelected(size.Value)
			}
		}
	} else if controller == groupTXT && target == groupSTT {
		if c.Controls.TxtSize != nil && c.Controls.TxtSize.GetSelected() != nil && c.Controls.STTModelSize != nil {
			size := c.Controls.TxtSize.GetSelected()
			if c.Controls.STTModelSize.ContainsEntry(size, CustomWidget.CompareValue) {
				c.Controls.STTModelSize.SetSelected(size.Value)
			}
		}
	}
	var devSource, devTarget *CustomWidget.TextValueSelect
	switch controller {
	case groupSTT:
		devSource = c.Controls.STTDevice
	case groupTXT:
		devSource = c.Controls.TxtDevice
	case groupOCR:
		devSource = c.Controls.OCRDevice
	}
	switch target {
	case groupSTT:
		devTarget = c.Controls.STTDevice
	case groupTXT:
		devTarget = c.Controls.TxtDevice
	case groupOCR:
		devTarget = c.Controls.OCRDevice
	}
	if devSource != nil && devTarget != nil && devSource.GetSelected() != nil && devTarget.ContainsEntry(devSource.GetSelected(), CustomWidget.CompareValue) {
		devTarget.SetSelected(devSource.GetSelected().Value)
	}
	var precSource, precTarget *CustomWidget.TextValueSelect
	switch controller {
	case groupSTT:
		precSource = c.Controls.STTPrecision
	case groupTXT:
		precSource = c.Controls.TxtPrecision
	case groupOCR:
		precSource = c.Controls.OCRPrecision
	}
	switch target {
	case groupSTT:
		precTarget = c.Controls.STTPrecision
	case groupTXT:
		precTarget = c.Controls.TxtPrecision
	case groupOCR:
		precTarget = c.Controls.OCRPrecision
	}
	if precSource != nil && precTarget != nil && precSource.GetSelected() != nil && precTarget.ContainsEntry(precSource.GetSelected(), CustomWidget.CompareValue) {
		precTarget.SetSelected(precSource.GetSelected().Value)
	}
	c.disableGroup(target)
	if c.MirrorLocked == nil {
		c.MirrorLocked = map[string]bool{}
	}
	c.MirrorLocked[target] = true
}

func (c *Coordinator) HandleMultiModalAllSync() {
	prev := c.InProgrammaticUpdate
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = prev }()
	sttType := c.getGroupType(groupSTT)
	txtType := c.getGroupType(groupTXT)
	ocrType := c.getGroupType(groupOCR)
	if c.MirrorLocked == nil {
		c.MirrorLocked = map[string]bool{}
	}
	mirroredNow := map[string]bool{}
	for modelType, caps := range multiModalCapabilities {
		groups := make([]string, 0, 3)
		if caps.STT && sttType == modelType {
			groups = append(groups, groupSTT)
		}
		if caps.TXT && txtType == modelType {
			groups = append(groups, groupTXT)
		}
		if caps.OCR && ocrType == modelType {
			groups = append(groups, groupOCR)
		}
		if len(groups) <= 1 {
			continue
		}
		controller := groupSTT
		if groups[0] != groupSTT {
			hasSTT := false
			for _, g := range groups {
				if g == groupSTT {
					hasSTT = true
					break
				}
			}
			if !hasSTT {
				controller = groupTXT
				hasTXT := false
				for _, g := range groups {
					if g == groupTXT {
						hasTXT = true
						break
					}
				}
				if !hasTXT {
					controller = groupOCR
				}
			}
		}
		for _, g := range groups {
			if g == controller {
				continue
			}
			c.mirrorFromTo(controller, g)
			mirroredNow[g] = true
		}
	}
	if sttType == "" {
		c.disableGroup(groupSTT)
		c.MirrorLocked[groupSTT] = false
	}
	if txtType == "" {
		c.disableGroup(groupTXT)
		c.MirrorLocked[groupTXT] = false
	}
	if ocrType == "" {
		c.disableGroup(groupOCR)
		c.MirrorLocked[groupOCR] = false
	}
	if sttType != "" && c.MirrorLocked[groupSTT] && !mirroredNow[groupSTT] {
		c.enableGroup(groupSTT)
		c.MirrorLocked[groupSTT] = false
	}
	if txtType != "" && c.MirrorLocked[groupTXT] && !mirroredNow[groupTXT] {
		c.enableGroup(groupTXT)
		c.MirrorLocked[groupTXT] = false
	}
	if ocrType != "" && c.MirrorLocked[groupOCR] && !mirroredNow[groupOCR] {
		c.enableGroup(groupOCR)
		c.MirrorLocked[groupOCR] = false
	}
}

func (c *Coordinator) ApplySTTTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	c.applyTypeChangeGeneric(
		modelType,
		c.Controls.STTModelSize,
		c.Controls.STTPrecision,
		c.Controls.STTDevice,
		STTModelOptions,
		STTPrecisionOptions,
		nil, // default device options
		"Whisper",
		false,
	)
	c.promptMultiModalAdoption(modelType, groupSTT)
}

func (c *Coordinator) ApplyTXTTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	c.applyTypeChangeGeneric(
		modelType,
		c.Controls.TxtSize,
		c.Controls.TxtPrecision,
		c.Controls.TxtDevice,
		TXTSizeOptions,
		TXTPrecisionOptions,
		nil, // default device options
		"TxtTranslator",
		false,
	)
	c.promptMultiModalAdoption(modelType, groupTXT)
}

func (c *Coordinator) ApplyTTSTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	c.applyTypeChangeGeneric(
		modelType,
		nil,
		nil,
		c.Controls.TTSDevice,
		nil,
		nil,
		nil, // default device options
		"ttsType",
		true, // fixed float32 precision for memory estimation
	)
}

func (c *Coordinator) ApplyOCRTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	c.applyTypeChangeGeneric(
		modelType,
		nil,
		c.Controls.OCRPrecision,
		c.Controls.OCRDevice,
		nil,
		OCRPrecisionOptions,
		OCRDeviceOptions,
		"ocrType",
		false,
	)
	c.promptMultiModalAdoption(modelType, groupOCR)
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// applyTypeChangeGeneric centralizes option population, enable/disable logic, and memory estimation.
func (c *Coordinator) applyTypeChangeGeneric(
	modelType string,
	sizeSel, precSel, devSel *CustomWidget.TextValueSelect,
	getSizeOptions func(string) ([]TVO, int, bool),
	getPrecisionOptions func(string) ([]TVO, bool),
	getDeviceOptions func(string) []TVO,
	aiModel string,
	setFixedFloat32Precision bool,
) {
	// Handle empty model: disable relevant controls and update memory with placeholder
	if modelType == "" {
		if sizeSel != nil {
			sizeSel.Disable()
		}
		if precSel != nil {
			precSel.Disable()
		}
		if devSel != nil {
			devSel.Disable()
		}
		AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: aiModel, AIModelType: "-"}
		if setFixedFloat32Precision {
			AIModel.Precision = Hardwareinfo.Float32
		}
		AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)
		c.HandleMultiModalAllSync()
		return
	}

	// Device options
	if devSel != nil {
		var devOpts []TVO
		if getDeviceOptions != nil {
			devOpts = getDeviceOptions(modelType)
		} else {
			devOpts = DefaultDeviceOptions()
		}
		c.SetOptionsWithFallback(devSel, devOpts)
		if len(devOpts) == 1 {
			devSel.SetSelected(devOpts[0].Value)
			devSel.Disable()
		} else {
			devSel.Enable()
		}
	}

	// Size options
	if sizeSel != nil && getSizeOptions != nil {
		if sOpts, defIdx, sEnable := getSizeOptions(modelType); sOpts != nil {
			c.SetOptionsWithFallback(sizeSel, sOpts)
			if sizeSel.GetSelected() == nil && len(sOpts) > defIdx {
				sizeSel.SetSelectedIndex(defIdx)
			}
			if !sEnable {
				sizeSel.Disable()
			} else if len(sOpts) == 1 {
				sizeSel.SetSelected(sOpts[0].Value)
				sizeSel.Disable()
			} else {
				sizeSel.Enable()
			}
		} else {
			sizeSel.Disable()
		}
	} else if sizeSel != nil {
		sizeSel.Disable()
	}

	// Precision options
	if precSel != nil && getPrecisionOptions != nil {
		if pOpts, pEnable := getPrecisionOptions(modelType); pOpts != nil {
			c.SetOptionsWithFallback(precSel, pOpts)
			if !pEnable {
				precSel.Disable()
			} else if len(pOpts) == 1 {
				precSel.SetSelected(pOpts[0].Value)
				precSel.Disable()
			} else {
				precSel.Enable()
			}
		} else {
			precSel.Disable()
		}
	} else if precSel != nil {
		precSel.Disable()
	}

	// Memory estimation
	AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: aiModel, AIModelType: firstNonEmpty(modelType, "-")}
	if setFixedFloat32Precision {
		AIModel.Precision = Hardwareinfo.Float32
	}
	AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)

	// Central sync (mirrors device/precision for multi-modal combos and locks targets)
	c.HandleMultiModalAllSync()
}

// promptMultiModalAdoption asks to reuse multi-modal models across groups except the one that initiated the change.
func (c *Coordinator) promptMultiModalAdoption(modelType string, excludeGroup string) {
	if c.SuppressPrompts || !MultiModalModels()[modelType] {
		return
	}
	caps, ok := multiModalCapabilities[modelType]
	if !ok {
		return
	}
	if caps.STT && excludeGroup != groupSTT && c.Controls.STTType != nil && c.getSelectedType(c.Controls.STTType) != modelType {
		c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Speech-to-Text as well?", func() {
			c.Controls.STTType.SetSelected(modelType)
			c.ApplySTTTypeChange(modelType)
		})
	}
	if caps.TXT && excludeGroup != groupTXT && c.Controls.TxtType != nil && c.getSelectedType(c.Controls.TxtType) != modelType {
		c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Text-Translation as well?", func() {
			c.Controls.TxtType.SetSelected(modelType)
			c.ApplyTXTTypeChange(modelType)
		})
	}
	if caps.OCR && excludeGroup != groupOCR && c.Controls.OCRType != nil && c.getSelectedType(c.Controls.OCRType) != modelType {
		c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Image-to-Text as well?", func() {
			c.Controls.OCRType.SetSelected(modelType)
			c.ApplyOCRTypeChange(modelType)
		})
	}
}
