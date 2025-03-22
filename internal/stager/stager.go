package stager

type Stager struct {
	stage     Stage
	lastStage Stage
	onChange  func(oldStage, newStage Stage)
}

type Stage int

const (
	GameStage Stage = iota
	MainMenuStage
	MenuStage
	SettingsStage
	Scene1Stage
)

func (stage Stage) String() string {
	switch stage {
	case GameStage:
		return "GameStage"
	case MainMenuStage:
		return "MainMenuStage"
	case MenuStage:
		return "MenuStage"
	case SettingsStage:
		return "SettingsStage"
	case Scene1Stage:
		return "Scene1Stage"
	}
	return ""
}

func New() *Stager {
	return &Stager{
		stage: MainMenuStage,
	}
}

func (stager *Stager) SetStage(newStage Stage) {
	stager.lastStage = stager.stage
	stager.stage = newStage
	if stager.onChange != nil {
		stager.onChange(stager.lastStage, newStage)
	}
}

func (stager *Stager) Stage() Stage {
	return stager.stage
}

func (stager *Stager) SetOnChange(onChange func(oldStage, newStage Stage)) {
	stager.onChange = onChange
}

func (stager *Stager) RecoveryLastStage() {
	stager.stage, stager.lastStage = stager.lastStage, stager.stage
}
