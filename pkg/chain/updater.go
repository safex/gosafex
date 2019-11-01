package chain

import (
	"time"
)

const UpdateCycleTime = 200
const CheckCycleTime = 30000
const DaemonErrorTime = 2000
const BlocksPerCycle = 500

func (w *Wallet) Rescan(accountName string) {
	select {
	case w.rescan <- accountName:
		w.logger.Debugf("[Updater] Rescanning for: %s", accountName)
	default:
		w.logger.Infof("[Updater] Error communicating with updater for rescan")
	}
}

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
	if w.rescanning != "" {
		return "Rescanning"
	}
	if w.syncing {
		return "Syncing"
	}
	return "Up-to-date"
}

func (w *Wallet) runUpdater() {
	if state := w.UpdaterStatus(); state != "Up-to-date" {
		w.logger.Infof("[Updater] Already: %s", state)
		return
	}
	var bcHeight uint64
	for true {
		select {
		case rscan := <-w.rescan:
			w.rescanning = rscan
		case temp := <-w.update:
			if temp == true {
				w.updating = true
			}
			if temp == false {
				w.syncing = false
				w.updating = false
			}
		case <-w.quit:
			w.syncing = false
			w.logger.Infof("[Updater] Updater Service Down")
			return
		default:
		}
		if w.updating && w.rescanning != "" {
			var err error
			loadedHeight := w.GetLatestLoadedBlockHeight()
			scannedHeight := uint64(1)
			if scannedHeight >= loadedHeight-1 {
				w.logger.Infof("[Updater] There was no need to rescan!")
			}
			for scannedHeight < loadedHeight-1 {
				err, scannedHeight = w.rescanBlocks(w.rescanning, scannedHeight, BlocksPerCycle)
				if err != nil {
					w.logger.Errorf("[Updater] Error while rescanning: %s", err.Error())
					break
				}
				w.logger.Infof("[Updater] Rescanned up to block %v", scannedHeight)
			}
			w.unlockBalance(loadedHeight)
			w.rescanning = ""
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
				if !w.working {
					if w.GetLatestLoadedBlockHeight() < bcHeight-1 {
						prevHeight := w.GetLatestLoadedBlockHeight()
						w.logger.Debugf("[Updater] Known block: %d , bcHeight: %d", w.GetLatestLoadedBlockHeight(), bcHeight)
						if err := w.updateBlocks(BlocksPerCycle); err != nil {
							w.logger.Errorf("[Updater] %s", err.Error())
						}
						if prevHeight == w.GetLatestLoadedBlockHeight() {
							w.logger.Errorf("[Updater] Can't load blocks")
							w.StopUpdating()
						}
					} else {
						w.syncing = false
					}
				} else {
					w.logger.Debugf("[Updater] Local DB busy")
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
