package sectorstorage

import (
	"bytes"
	"encoding/json"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"golang.org/x/xerrors"
	"net/http"
	"os"
	"strings"
	"time"
)

type OfflineRemoteWorker struct {
	Hostname string `json:"hostname"`
	Socket   string `json:"socket"`
	Ts       int64 `json:"ts"`
}

type SectorStage uint8

const (
	AddPiece SectorStage = iota
	SealPreCommit1
	SealPreCommit2
	SealCommit1
	SealCommit2
	Finalize
)

type SectorProcess struct {
	SectorID uint64 `json:"sectorID"`
	Hostname string `json:"hostname"`
	SectorStage  uint8 `json:"sectorStage"`
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

func TriggerSectorProcess(hostname string, sectorStage SectorStage, sectorID abi.SectorNumber, success bool) error {
	url := os.Getenv("WEBHOOK_SECTOR_PROCESS")
	if url == "" {
		return xerrors.Errorf("can not find WEBHOOK_SECTOR_PROCESS environment variable")
	}

	data := &SectorProcess{uint64(sectorID), hostname, uint8(sectorStage), success, makeTimestamp()}

	jsonValue, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return TriggerWebHook(url, jsonValue)
}

type rpcReq struct {
	JSONRPC string `json:"jsonrpc"`
	Method string `json:"method"`
	Id int `json:"id"`
	Params []interface{} `json:"params"`
}

func TriggerNextPledge() error {
	api := os.Getenv("MINER_API")

	if api == "" || !strings.Contains(api, "#") {
		return xerrors.Errorf("can not find MINER_API environment variable")
	}

	rawAPI := strings.Split(api, "#")
	url := rawAPI[0]
	token := "Bearer " + rawAPI[1]

	time.Sleep(10 * time.Second)

	data := []byte(`{"jsonrpc": "2.0", "method": "Filecoin.PledgeSector", "id": 1, "params": [] }`)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	if err != nil {
		log.Fatal("Error reading request. ", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", token)

	// Set client timeout
	client := &http.Client{Timeout: time.Second * 10}

	// Send request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}