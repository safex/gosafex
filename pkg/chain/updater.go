package chain

import (
	"time"
)

const UpdateCycleTime = 10
const CheckCycleTime = 500
const DaemonErrorTime = 2000
const BlocksPerCycle = 500

func (w *Wallet) StopUpdating() {
	select {
	case w.update <- false:
		w.logger.Infof("[Updater] Stop updating")
	default:
		w.logger.Infof("[Updater] Can't stop updating")
	}
}

func (w *Wallet) BeginUpdating() {
	w.logger.Infof("[Updater] Starting the updater service")
	go w.runUpdater()
	time.Sleep(1 * time.Second)
}

func (w *Wallet) KillUpdating() {
	w.logger.Infof("[Updater] Killing the updater service")
	select {
	case w.quit <- true:
	default:
	}
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
		case <-w.quit:
			w.syncing = false
			w.logger.Infof("[Updater] Updater Service Down")
			return
		default:
		}
		if !w.syncing {
			loadedHeight := w.GetLatestLoadedBlockHeight()
			info, err := w.client.GetDaemonInfo()
			if err != nil {
				time.Sleep(DaemonErrorTime * time.Millisecond)
				continue
			}
			bcHeight = info.Height
			if loadedHeight < bcHeight-1 {
				w.syncing = true
			}
			w.logger.Debugf("[Updater] Known block: %d", loadedHeight)
			time.Sleep(CheckCycleTime * time.Millisecond)

		} else {
			if w.GetLatestLoadedBlockHeight() < bcHeight-1 {
				w.logger.Debugf("[Updater] Known block: %d , bcHeight: %d", w.GetLatestLoadedBlockHeight(), bcHeight)
				w.UpdateBlock(BlocksPerCycle)
			} else {
				w.syncing = false
			}
			time.Sleep(UpdateCycleTime * time.Millisecond)
		}
	}
}
