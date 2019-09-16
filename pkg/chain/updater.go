package chain

import (
	"time"
)

const UpdateCycleTime = 50
const CheckCycleTime = 5000
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
	w.updating = true
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
	return "Up-to-date"
}

func (w *Wallet) runUpdater() {
	var bcHeight uint64
	for true {
		select {
		case temp := <-w.update:
			if temp == true {
				w.updating = true
			}
			if temp == false {
				w.updating = false
			}
		case <-w.quit:
			w.syncing = false
			w.logger.Infof("[Updater] Updater Service Down")
			return
		default:
		}
		if w.updating {
			if !w.syncing {
				loadedHeight := w.GetLatestLoadedBlockHeight()
				info, err := w.client.GetDaemonInfo()
				if err != nil {
					w.logger.Errorf("[Updater] Can't connect to daemon")
					time.Sleep(DaemonErrorTime * time.Millisecond)
					continue
				}
				w.latestInfo = &info
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
				info, err := w.client.GetDaemonInfo()
				if err != nil {
					w.logger.Debugf("[Updater] Unexpected error in client syncing")
					time.Sleep(UpdateCycleTime * time.Millisecond)
					continue
				}
				w.latestInfo = &info

				time.Sleep(UpdateCycleTime * time.Millisecond)
			}
		} else {
			time.Sleep(CheckCycleTime * time.Millisecond)
		}
	}
}
