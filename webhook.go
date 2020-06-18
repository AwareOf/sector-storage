package sectorstorage

import (
	"bytes"
	"encoding/json"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"golang.org/x/xerrors"
	"net/http"
	"os"
	"time"
)

type OfflineRemoteWorker struct {
	Hostname string `json:"hostname"`
	Socket   string `json:"socket"`
	Ts       int64 `json:"ts"`
}

type Process uint8

const (
	AddPiece Process = iota
	SealPreCommit1
	SealPreCommit2
	SealCommit1
	SealCommit2
	Finalize
)

type SectorProcess struct {
	SectorID uint64 `json:"sectorID"`
	Hostname string `json:"hostname"`
	Process  uint8 `json:"process"`
	Success  bool `json:"success"`
	Ts       int64 `json:"ts"`
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func TriggerWebHook(url string, data []byte) error {
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func TriggerWorkerOffline(hostname string, socket string) error {
	url := os.Getenv("WEBHOOK_WORKER_OFFLINE")
	if url == "" {
		return xerrors.Errorf("can not find $WEBHOOK_WORKER_OFFLINE")
	}

	data := &OfflineRemoteWorker{hostname, socket, makeTimestamp()}

	jsonValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return TriggerWebHook(url, jsonValue)
}

func TriggerSectorProcess(hostname string, process Process, sectorID abi.SectorNumber, success bool) error {
	url := os.Getenv("WEBHOOK_SECTOR_PROCESS")
	if url == "" {
		return xerrors.Errorf("can not find WEBHOOK_SECTOR_PROCESS environment variable")
	}

	data := &SectorProcess{uint64(sectorID), hostname, uint8(process), success, makeTimestamp()}

	jsonValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return TriggerWebHook(url, jsonValue)
}
