package Pages

import (
	"whispering-tiger-ui/CustomWidget"
	"whispering-tiger-ui/Utilities/Hardwareinfo"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/lang"
	"fyne.io/fyne/v2/widget"
)

// ProfileControls groups commonly used widgets of the Profiles form.
// It lets us apply cross-field logic in one place and keep Profiles.go tidy.
type ProfileControls struct {
	// STT
	STTType      *CustomWidget.TextValueSelect
	STTDevice    *CustomWidget.TextValueSelect
	STTPrecision *CustomWidget.TextValueSelect
	STTModelSize *CustomWidget.TextValueSelect

	// Text translator
	TxtType      *CustomWidget.TextValueSelect
	TxtDevice    *CustomWidget.TextValueSelect
	TxtPrecision *CustomWidget.TextValueSelect
	TxtSize      *CustomWidget.TextValueSelect

	// OCR
	OCRType      *CustomWidget.TextValueSelect
	OCRDevice    *CustomWidget.TextValueSelect
	OCRPrecision *CustomWidget.TextValueSelect

	// TTS
	TTSType   *CustomWidget.TextValueSelect
	TTSDevice *CustomWidget.TextValueSelect
}

// Coordinator centralizes dynamic behaviors: enabling/disabling, option fallback, and multi-modal sync.
type Coordinator struct {
	Controls          *ProfileControls
	IsLoadingSettings *bool
	ComputeCapability float32
	CPUMemoryBar      *widget.ProgressBar
	GPUMemoryBar      *widget.ProgressBar
	TotalGPUMemoryMiB int64
	// Verhindert Event-Kaskaden bei programmgesteuerten SetSelected-Aufrufen
	InProgrammaticUpdate bool
	// Unterdrückt Bestätigungs-Dialoge während programmatischer Übernahmen
	SuppressPrompts bool
	// Merkt sich, welche Gruppen aufgrund Multi-Modal-Mirroring gesperrt (disabled) sind
	MirrorLocked map[string]bool
}

// getParentWindow returns the last available window to use as dialog parent.
func (c *Coordinator) getParentWindow() fyne.Window {
	wins := fyne.CurrentApp().Driver().AllWindows()
	if len(wins) == 0 {
		return nil
	}
	return wins[len(wins)-1]
}

// SetOptionsWithFallback sets options on a TextValueSelect and ensures the selection remains valid.
// If the current selection is not available, it falls back to the first option (if any).
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
		// Preserve current selection
		sel.SetSelected(current.Value)
	}
	sel.Refresh()
}

// Enable enables provided widgets
func (c *Coordinator) Enable(widgets ...fyne.Disableable) {
	for _, w := range widgets {
		if w != nil {
			w.Enable()
		}
	}
}

// Disable disables provided widgets
func (c *Coordinator) Disable(widgets ...fyne.Disableable) {
	for _, w := range widgets {
		if w != nil {
			w.Disable()
		}
	}
}

// IsMultiModalModelPair reports if both selects use the same multi-modal model type.
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

// SyncMultiModalTargets sets target widgets to mirror the source and disables them.
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

// HandleMultiModalModelSync applies synchronization/enable rules for multi-modal model pairs.
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
	// Two distinct, non-empty models: ensure target widgets are enabled.
	if modelSelect1.GetSelected() != nil && modelSelect1.GetSelected().Value != "" &&
		modelSelect2.GetSelected() != nil && modelSelect2.GetSelected().Value != "" {
		c.Enable(targetSize, targetPrecision, targetDevice)
	}
}

// ConfirmOption proposes a confirmation dialog when a setting may want to change other fields.
func (c *Coordinator) ConfirmOption(titleKey, messageKey string, onYes func()) {
	if c.IsLoadingSettings != nil && *c.IsLoadingSettings {
		return
	}
	parent := c.getParentWindow()
	dialog.NewConfirm(lang.L(titleKey), lang.L(messageKey), func(b bool) {
		if !b {
			return
		}
		// führe Änderungen auf dem UI-Thread aus
		fyne.Do(func() {
			// Unterdrücke während der bestätigten Übernahme weitere Prompts,
			// um Reentranz/Nesting zu vermeiden
			prevProg := c.InProgrammaticUpdate
			prevPrompts := c.SuppressPrompts
			c.InProgrammaticUpdate = true
			c.SuppressPrompts = true
			onYes()
			// Nach der eigentlichen Änderung alle Sync-Regeln anwenden,
			// weiterhin ohne weitere Prompts
			c.HandleMultiModalAllSync()
			// Flags zurücksetzen
			c.SuppressPrompts = prevPrompts
			c.InProgrammaticUpdate = prevProg
		})
	}, parent).Show()
}

// EnsurePrecisionDeviceCompatibility shows informational prompts when precision/device pairs are known to be problematic.
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

// Multi-modal capability map
// Key: model type; Value: capabilities across groups
var multiModalCapabilities = map[string]struct{ STT, TXT, OCR bool }{
	"seamless_m4t": {STT: true, TXT: true, OCR: false},
	"voxtral":      {STT: true, TXT: true, OCR: false},
	"phi4":         {STT: true, TXT: true, OCR: true},
}

// group identifiers for internal logic
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

// enable/disable helpers per group
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

// mirror settings from controller group to target group
func (c *Coordinator) mirrorFromTo(controller, target string) {
	// size
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
	// device
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
	// precision
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
	// finally disable the target group controls
	c.disableGroup(target)
	if c.MirrorLocked == nil {
		c.MirrorLocked = map[string]bool{}
	}
	c.MirrorLocked[target] = true
}

// HandleMultiModalAllSync evaluates current selections and applies synchronization and enable/disable rules
// Priority: STT > TXT > OCR
func (c *Coordinator) HandleMultiModalAllSync() {
	// Unterdrücke OnChanged-Kaskaden während der Synchronisierung
	prev := c.InProgrammaticUpdate
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = prev }()

	// collect selected types
	sttType := c.getGroupType(groupSTT)
	txtType := c.getGroupType(groupTXT)
	ocrType := c.getGroupType(groupOCR)

	if c.MirrorLocked == nil {
		c.MirrorLocked = map[string]bool{}
	}

	// Track, welche Gruppen in diesem Durchlauf gespiegelt (und dadurch gesperrt) werden
	mirroredNow := map[string]bool{}

	// handle each known multimodal type
	for modelType, caps := range multiModalCapabilities {
		// find which groups currently selected this type
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
		// determine controller by priority
		controller := groupSTT
		if groups[0] != groupSTT {
			// check presence
			hasSTT := false
			for _, g := range groups {
				if g == groupSTT {
					hasSTT = true
					break
				}
			}
			if !hasSTT {
				// next priority TXT
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
		// mirror to others
		for _, g := range groups {
			if g == controller {
				continue
			}
			c.mirrorFromTo(controller, g)
			mirroredNow[g] = true
		}
	}

	// Gruppen, die leer (Disabled) sind, bleiben deaktiviert und werden nicht freigegeben
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

	// Ehemals durch Mirroring gesperrte Gruppen wieder freigeben, sofern sie nicht mehr gespiegelt werden
	// und nicht Disabled sind. So reaktivieren wir nur die Sperre aus dem Mirroring, nicht modell-spezifische Sperren.
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

// ApplySTTTypeChange updates STT-related controls based on selected STT model type.
// It also prompts for multi-modal adoption and refreshes memory estimation.
func (c *Coordinator) ApplySTTTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	sttSize := c.Controls.STTModelSize
	sttPrec := c.Controls.STTPrecision
	sttDev := c.Controls.STTDevice

	// Empty: disable all related controls and update memory, then return
	if modelType == "" {
		c.Disable(sttSize, sttPrec, sttDev)
		AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "Whisper", AIModelType: "-"}
		AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)
		c.HandleMultiModalAllSync()
		return
	}

	// Enable by default
	c.Enable(sttSize, sttPrec, sttDev)

	// Device options (currently generic)
	c.SetOptionsWithFallback(sttDev, DefaultDeviceOptions())

	// Model size options
	if opts, defIdx, enable := STTModelOptions(modelType); enable {
		c.SetOptionsWithFallback(sttSize, opts)
		if sttSize.GetSelected() == nil && len(opts) > defIdx {
			sttSize.SetSelectedIndex(defIdx)
		}
	} else {
		sttSize.Disable()
	}

	// Precision options
	if pOpts, pEnable := STTPrecisionOptions(modelType); pEnable {
		c.SetOptionsWithFallback(sttPrec, pOpts)
	} else {
		if modelType == "nemo_canary" {
			// fixed float32
			c.SetOptionsWithFallback(sttPrec, []CustomWidget.TextValueOption{{Text: "float32 " + lang.L("precision"), Value: "float32"}})
			sttPrec.Disable()
		} else {
			sttPrec.Disable()
		}
	}

	// Special handling
	switch modelType {
	case "speech_t5":
		sttPrec.Disable()
		sttSize.Disable()
	case "phi4":
		// size fixed to large
		c.SetOptionsWithFallback(sttSize, []CustomWidget.TextValueOption{{Text: "Large", Value: "large"}})
		sttSize.Disable()
		// precision limited
		c.SetOptionsWithFallback(sttPrec, []CustomWidget.TextValueOption{
			{Text: "float32 " + lang.L("precision"), Value: "float32"},
			{Text: "float16 " + lang.L("precision"), Value: "float16"},
			{Text: "bfloat16 " + lang.L("precision"), Value: "bfloat16"},
		})
	}

	// Multi-modal adoption prompts
	if !c.SuppressPrompts && MultiModalModels()[modelType] {
		// Offer TXT adoption
		if c.Controls.TxtType != nil && c.getSelectedType(c.Controls.TxtType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Text-Translation as well?", func() {
				c.Controls.TxtType.SetSelected(modelType)
				c.ApplyTXTTypeChange(modelType)
			})
		}
		// Offer OCR adoption (only for models supporting OCR)
		if multiModalCapabilities[modelType].OCR && c.Controls.OCRType != nil && c.getSelectedType(c.Controls.OCRType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Image-to-Text as well?", func() {
				c.Controls.OCRType.SetSelected(modelType)
				c.ApplyOCRTypeChange(modelType)
			})
		}
	}

	// Memory consumption update
	AIModel := Hardwareinfo.ProfileAIModelOption{
		AIModel:     "Whisper",
		AIModelType: firstNonEmpty(modelType, "-"),
	}
	AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)

	// Sync all multimodal groups
	c.HandleMultiModalAllSync()
}

// ApplyTXTTypeChange updates TXT-related controls based on selected model type.
func (c *Coordinator) ApplyTXTTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	txtDev := c.Controls.TxtDevice
	txtPrec := c.Controls.TxtPrecision
	txtSize := c.Controls.TxtSize

	// Empty: disable all related controls and update memory, then return
	if modelType == "" {
		c.Disable(txtDev, txtPrec, txtSize)
		AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "TxtTranslator", AIModelType: "-"}
		AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)
		c.HandleMultiModalAllSync()
		return
	}

	c.Enable(txtDev, txtPrec, txtSize)
	// Device options
	c.SetOptionsWithFallback(txtDev, TXTDeviceOptions(modelType))
	// Geräteeinschränkungen pro Modelltyp (Beispiele)
	switch modelType {
	case "NLLB200":
		// Original NLLB: meist CPU-only im Setup → CUDA deaktivieren
		if txtDev != nil {
			txtDev.SetSelected("cpu")
			txtDev.Disable()
		}
	}
	// Precision options
	if pOpts, pEnable := TXTPrecisionOptions(modelType); pEnable {
		c.SetOptionsWithFallback(txtPrec, pOpts)
	} else {
		c.Disable(txtPrec)
	}
	// Size options
	if sOpts, defIdx, sEnable := TXTSizeOptions(modelType); sEnable {
		c.SetOptionsWithFallback(txtSize, sOpts)
		if txtSize.GetSelected() == nil && len(sOpts) > defIdx {
			txtSize.SetSelectedIndex(defIdx)
		}
	} else {
		c.Disable(txtSize)
	}

	// Memory consumption update
	AIModel := Hardwareinfo.ProfileAIModelOption{
		AIModel:     "TxtTranslator",
		AIModelType: firstNonEmpty(modelType, "-"),
	}
	AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)

	// Sync
	c.HandleMultiModalAllSync()

	// Multi-modal adoption prompts (symmetrical to STT)
	if !c.SuppressPrompts && MultiModalModels()[modelType] {
		// Offer STT adoption
		if c.Controls.STTType != nil && c.getSelectedType(c.Controls.STTType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Speech-to-Text as well?", func() {
				c.Controls.STTType.SetSelected(modelType)
				c.ApplySTTTypeChange(modelType)
			})
		}
		// Offer OCR adoption when supported
		if multiModalCapabilities[modelType].OCR && c.Controls.OCRType != nil && c.getSelectedType(c.Controls.OCRType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Image-to-Text as well?", func() {
				c.Controls.OCRType.SetSelected(modelType)
				c.ApplyOCRTypeChange(modelType)
			})
		}
	}
}

// ApplyTTSTypeChange updates TTS device options and memory consumption.
func (c *Coordinator) ApplyTTSTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	ttsDev := c.Controls.TTSDevice
	if ttsDev == nil {
		return
	}
	if modelType != "" {
		c.Enable(ttsDev)
		c.SetOptionsWithFallback(ttsDev, TTSDeviceOptions(modelType))
	} else {
		c.Disable(ttsDev)
	}
	AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ttsType", AIModelType: firstNonEmpty(modelType, "-"), Precision: Hardwareinfo.Float32}
	AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)
}

// ApplyOCRTypeChange updates OCR enable/disable and sync with STT/TXT when equal.
func (c *Coordinator) ApplyOCRTypeChange(modelType string) {
	if c == nil || c.Controls == nil {
		return
	}
	c.InProgrammaticUpdate = true
	defer func() { c.InProgrammaticUpdate = false }()
	oDev := c.Controls.OCRDevice
	oPrec := c.Controls.OCRPrecision
	c.Enable(oDev, oPrec)
	if modelType == "" {
		c.Disable(oDev, oPrec)
	} else if modelType == "easyocr" {
		if oDev != nil {
			oDev.SetSelected("cpu")
			oDev.Disable()
		}
		if oPrec != nil {
			oPrec.Disable()
		}
	} else if c.Controls.STTType != nil && c.getSelectedType(c.Controls.STTType) == modelType {
		// mirror from STT
		if oDev != nil && c.Controls.STTDevice != nil && c.Controls.STTDevice.GetSelected() != nil && oDev.ContainsEntry(c.Controls.STTDevice.GetSelected(), CustomWidget.CompareValue) {
			oDev.SetSelected(c.Controls.STTDevice.GetSelected().Value)
		}
		if oPrec != nil && c.Controls.STTPrecision != nil && c.Controls.STTPrecision.GetSelected() != nil && oPrec.ContainsEntry(c.Controls.STTPrecision.GetSelected(), CustomWidget.CompareValue) {
			oPrec.SetSelected(c.Controls.STTPrecision.GetSelected().Value)
		}
		c.Disable(oDev, oPrec)
	} else if c.Controls.TxtType != nil && c.getSelectedType(c.Controls.TxtType) == modelType {
		// mirror from TXT
		if oDev != nil && c.Controls.TxtDevice != nil && c.Controls.TxtDevice.GetSelected() != nil && oDev.ContainsEntry(c.Controls.TxtDevice.GetSelected(), CustomWidget.CompareValue) {
			oDev.SetSelected(c.Controls.TxtDevice.GetSelected().Value)
		}
		if oPrec != nil && c.Controls.TxtPrecision != nil && c.Controls.TxtPrecision.GetSelected() != nil && oPrec.ContainsEntry(c.Controls.TxtPrecision.GetSelected(), CustomWidget.CompareValue) {
			oPrec.SetSelected(c.Controls.TxtPrecision.GetSelected().Value)
		}
		c.Disable(oDev, oPrec)
	}

	// Memory
	AIModel := Hardwareinfo.ProfileAIModelOption{AIModel: "ocrType", AIModelType: firstNonEmpty(modelType, "-")}
	AIModel.CalculateMemoryConsumption(c.CPUMemoryBar, c.GPUMemoryBar, c.TotalGPUMemoryMiB)
	c.HandleMultiModalAllSync()

	// Multi-modal adoption prompts when OCR model is multi-modal
	if !c.SuppressPrompts && MultiModalModels()[modelType] {
		// Offer STT adoption if supported
		if multiModalCapabilities[modelType].STT && c.Controls.STTType != nil && c.getSelectedType(c.Controls.STTType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Speech-to-Text as well?", func() {
				c.Controls.STTType.SetSelected(modelType)
				c.ApplySTTTypeChange(modelType)
			})
		}
		// Offer TXT adoption if supported
		if multiModalCapabilities[modelType].TXT && c.Controls.TxtType != nil && c.getSelectedType(c.Controls.TxtType) != modelType {
			c.ConfirmOption("Usage of Multi-Modal Model.", "Use Multi-Modal model for Text-Translation as well?", func() {
				c.Controls.TxtType.SetSelected(modelType)
				c.ApplyTXTTypeChange(modelType)
			})
		}
	}
}

// firstNonEmpty returns a if non-empty, otherwise b
func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
