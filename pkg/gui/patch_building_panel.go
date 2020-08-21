package gui

import (
	"github.com/jesseduffield/gocui"
	"github.com/jesseduffield/lazygit/pkg/utils"
)

func (gui *Gui) refreshPatchBuildingPanel(selectedLineIdx int) error {
	if !gui.GitCommand.PatchManager.Active() {
		return gui.handleEscapePatchBuildingPanel()
	}

	gui.splitMainPanel(true)

	gui.getMainView().Title = "Patch"
	gui.getSecondaryView().Title = "Custom Patch"

	// get diff from commit file that's currently selected
	commitFile := gui.getSelectedCommitFile()
	if commitFile == nil {
		gui.renderString("commitFiles", gui.Tr.SLocalize("NoCommiteFiles"))
		return nil
	}

	diff, err := gui.GitCommand.ShowCommitFile(commitFile.Parent, commitFile.Name, true)
	if err != nil {
		return err
	}

	secondaryDiff := gui.GitCommand.PatchManager.RenderPatchForFile(commitFile.Name, true, false, true)
	if err != nil {
		return err
	}

	empty, err := gui.refreshLineByLinePanel(diff, secondaryDiff, false, selectedLineIdx)
	if err != nil {
		return err
	}

	if empty {
		return gui.handleEscapePatchBuildingPanel()
	}

	return nil
}

func (gui *Gui) handleToggleSelectionForPatch(g *gocui.Gui, v *gocui.View) error {
	state := gui.State.Panels.LineByLine

	toggleFunc := gui.GitCommand.PatchManager.AddFileLineRange
	filename := gui.getSelectedCommitFileName()
	includedLineIndices := gui.GitCommand.PatchManager.GetFileIncLineIndices(filename)
	currentLineIsStaged := utils.IncludesInt(includedLineIndices, state.SelectedLineIdx)
	if currentLineIsStaged {
		toggleFunc = gui.GitCommand.PatchManager.RemoveFileLineRange
	}

	// add range of lines to those set for the file
	commitFile := gui.getSelectedCommitFile()
	if commitFile == nil {
		gui.renderString("commitFiles", gui.Tr.SLocalize("NoCommiteFiles"))
		return nil
	}

	toggleFunc(commitFile.Name, state.FirstLineIdx, state.LastLineIdx)

	if err := gui.refreshCommitFilesView(); err != nil {
		return err
	}

	if err := gui.refreshPatchBuildingPanel(-1); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) handleEscapePatchBuildingPanel() error {
	gui.handleEscapeLineByLinePanel()

	if gui.GitCommand.PatchManager.IsEmpty() {
		gui.GitCommand.PatchManager.Reset()
	}

	return gui.switchContext(gui.Contexts.CommitFiles.Context)
}

func (gui *Gui) refreshSecondaryPatchPanel() error {
	// TODO: swap out for secondaryPatchPanelUpdateOpts

	if gui.GitCommand.PatchManager.Active() {
		gui.splitMainPanel(true)
		secondaryView := gui.getSecondaryView()
		secondaryView.Highlight = true
		secondaryView.Wrap = false

		gui.g.Update(func(*gocui.Gui) error {
			gui.setViewContent(gui.getSecondaryView(), gui.GitCommand.PatchManager.RenderAggregatedPatchColored(false))
			return nil
		})
	} else {
		gui.splitMainPanel(false)
	}

	return nil
}

func (gui *Gui) secondaryPatchPanelUpdateOpts() *viewUpdateOpts {
	if gui.GitCommand.PatchManager.Active() {
		patch := gui.GitCommand.PatchManager.RenderAggregatedPatchColored(false)

		return &viewUpdateOpts{
			title:     "Custom Patch",
			noWrap:    true,
			highlight: true,
			task:      gui.createRenderStringWithoutScrollTask(patch),
		}
	}

	return nil
}
