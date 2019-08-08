package chain

import (
	"time"
)

const UpdateCycleTime = 10
const CheckCycleTime = 500
const DaemonErrorTime = 2000
const BlocksPerCycle = 500

func (w *Wallet) StartUpdating() {
	w.update <- true
}

func (w *Wallet) StopUpdating() {
	w.update <- false
}

func (w *Wallet) BeginUpdating() {
	go w.runUpdater()
}

func (w *Wallet) KillUpdating() {
	w.quit <- true
}

func (w *Wallet) UpdaterStatus() string {
	if w.syncing {
		return "Syncing"
	}
	if w.updating {
		return "Up to date"
	}
	return "Not updating"

}

func (w *Wallet) runUpdater() {
	var bcHeight uint64
	for true {
		select {
		case w.updating = <-w.update:
		case w.quitting = <-w.quit:
		default:
		}
		if w.quitting {
			break
		}
		if w.updating {
			if !w.syncing {
				loadedHeight := w.GetLatestLoadedBlockHeight()
				info, err := w.client.GetDaemonInfo()
				if err != nil {
					time.Sleep(DaemonErrorTime * time.Millisecond)
					continue
				}
				bcHeight = info.Height
				if loadedHeight != bcHeight {
					w.syncing = true
				}
				time.Sleep(CheckCycleTime * time.Millisecond)

			} else {
				if w.GetLatestLoadedBlockHeight() <= bcHeight-1 {
					w.UpdateBlock(BlocksPerCycle)
				} else {
					w.syncing = false
				}
				time.Sleep(UpdateCycleTime * time.Millisecond)

			}
		}
	}
}
