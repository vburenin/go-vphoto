package downloader

import (
	log "github.com/sirupsen/logrus"
	"github.com/vburenin/go-vphoto/authstate"
	"github.com/vburenin/go-vphoto/models"
	"github.com/vburenin/go-vphoto/settings"
	"io/ioutil"
	"os"
	"path"
	"sync"
	"time"
)
import "github.com/vburenin/nsync"

var workersPool = nsync.NewControlWaitGroup(100)

type PicDownloader struct {
	downloadMap   map[string]*models.MediaItem

	downloadQueue chan *models.MediaItem
	mu            *sync.Mutex
	picLock *nsync.NamedMutex
}


func NewPicDownloader() *PicDownloader {
	p := &PicDownloader{
		downloadMap: make(map[string]*models.MediaItem, 16384),
		downloadQueue: make(chan *models.MediaItem, 1024*1024),
		mu: &sync.Mutex{},
		picLock: nsync.NewNamedMutex(),
	}
	go p.downloadLoop()
	return p
}

var Downloader *PicDownloader

func InitDownloader() {
	Downloader = NewPicDownloader()
}

func (pd *PicDownloader) AddToDownload(m *models.MediaItem) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if ci, ok := pd.downloadMap[m.ID]; ok {
		m.Ready = ci.Ready
		return
	}
	m.Ready = make(chan bool)
	pd.downloadMap[m.ID] = m
	pd.downloadQueue <- m
}

var sentinelItem = &models.MediaItem{}

func picPath(picID string) string {
	return path.Join(settings.CacheLocation, picID)
}

func (pd *PicDownloader) retrievePicture(m *models.MediaItem) {
	pd.picLock.Lock(m.ID)
	defer pd.picLock.Unlock(m.ID)

	pd.mu.Lock()
	_, ok := pd.downloadMap[m.ID]
	pd.mu.Unlock()
	if !ok {
		return
	}

	// remove processed ID at the end.
	defer func() {
		pd.mu.Lock()
		if _, ok := pd.downloadMap[m.ID]; ok {
			delete(pd.downloadMap, m.ID)
		}
		pd.mu.Unlock()
	}()

	filePath := picPath(m.ID)
	defer close(m.Ready)
	s, err := os.Stat(filePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Errorf("could not download picture: %s", err)
			return
		}
	}

	if s != nil && s.Size() > 0 {
		return
	}

	c := authstate.GetHttpClient()
	url := m.BaseURL + "=d"
	resp, err := c.Get(url)
	if err != nil {
		log.Errorf("could not retrieve picture %s: %s", url, err)
	}
	defer resp.Body.Close()
	data, _ := ioutil.ReadAll(resp.Body)
	fh, err := os.Create(filePath)
	if err != nil {
		log.Errorf("could not store picture file %s: %s", filePath, err)
		return
	}
	defer fh.Close()
	_, err = fh.Write(data)
	if err != nil {
		log.Errorf("could not write down picture file %s: %s", filePath, err)
	}
	log.Infof("successfully downloaded picture: %s", m.ID)
}

func (pd *PicDownloader) downloadLoop() {
	for {
		item := pd.getNextToDownload()
		if item == sentinelItem {
			return
		}
		if item == nil {
			continue
		}
		workersPool.Do(func() {
			pd.retrievePicture(item)
		})
	}
}

func (pd *PicDownloader) GetPicNow(ID string) string {
	p := picPath(ID)

	pd.picLock.Lock(ID)
	if s, err := os.Stat(p); err == nil {
		if s.Size() > 0 {
			pd.picLock.Unlock(ID)
			return p
		}
	}
	pd.picLock.Unlock(ID)

	item := &models.MediaItem{
		ID: ID,
		Ready: make(chan bool),
	}
	pd.retrievePicture(item)
	select {
		case <-item.Ready:
			return picPath(ID)
		case <-time.After(time.Second * 300):
			return ""
	}
}

func (pd *PicDownloader) getNextToDownload() *models.MediaItem{
	select {
	case item := <- pd.downloadQueue:
		if item == nil {
			return sentinelItem
		}
		return item
	case <- time.After(time.Second):
		return nil
	}
}

func (pd *PicDownloader) downloadPicture(m *models.MediaItem) {
	pd.mu.Lock()
	defer pd.mu.Unlock()

	if ci, ok := pd.downloadMap[m.ID]; ok {
		m.Ready = ci.Ready
		return
	}
	m.Ready = make(chan bool)
}

