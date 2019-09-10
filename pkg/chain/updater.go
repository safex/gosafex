package chain

import (
	"time"
)

const UpdateCycleTime = 10
const CheckCycleTime = 500
const DaemonErrorTime = 2000
const BlocksPerCycle = 500

func (w *Wallet) StartUpdating() {
	select {
	case w.update <- true:
		w.logger.Infof("[Wallet] Start updating")
	default:
		w.logger.Infof("[Wallet] Can't start updating")
	}
}

func (w *Wallet) StopUpdating() {
	select {
	case w.update <- false:
		w.logger.Infof("[Wallet] Stop updating")
	default:
		w.logger.Infof("[Wallet] Can't stop updating")
	}
}

func (w *Wallet) BeginUpdating() {
	w.logger.Infof("[Wallet] Starting the updater service")
	w.StartUpdating()
	go w.runUpdater()
}

func (w *Wallet) KillUpdating() {
	w.logger.Infof("[Wallet] Killing the updater service")
	w.quit <- true
}

func (w *Wallet) UpdaterStatus() string {
	if w.syncing {
		return "Syncing"
	}
	if w.updating {
		return "Up-to-date"
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
