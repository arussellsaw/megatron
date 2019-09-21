package library

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/arussellsaw/megatron/domain"
	"github.com/arussellsaw/megatron/domain/subdb"
	"github.com/blevesearch/bleve"
	"github.com/google/uuid"
	"github.com/influxdata/influxdb/kit/errors"
	"github.com/monzo/slog"
)

type Index struct {
	index               bleve.Index
	mu                  sync.Mutex
	emu                 sync.Mutex
	episodesByCaptionID map[string]string
	captionsByID        map[string]domain.Caption
	episodesByID        map[string]domain.Episode
}

type SearchResult struct {
	EpisodeID  string         `json:"episode_id"`
	Caption    domain.Caption `json:"caption"`
	Confidence float64        `json:"confidence"`
}

func BuildIndex(ctx context.Context, dir string) (*Index, error) {
	params := map[string]string{
		"index": "megatron.bleve",
		"dir":   dir,
	}

	slog.Info(ctx, "Creating new index from %s at megatron.bleve", dir, params)

	mapping := bleve.NewIndexMapping()
	index, err := bleve.New("megatron.bleve", mapping)
	if err != nil {
		slog.Error(ctx, "Error creating index: %s", err, params)
		return nil, err
	}

	i := Index{
		index:               index,
		episodesByCaptionID: make(map[string]string),
		episodesByID:        make(map[string]domain.Episode),
		captionsByID:        make(map[string]domain.Caption),
	}

	err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		params["path"] = path
		if err != nil {
			slog.Error(ctx, "Error walking library: %s", params)
			return err
		}
		if info.IsDir() {
			return nil
		}
		switch path[len(path)-3:] {
		case "mp4", "mkv", "avi", "m4v":
			slog.Info(ctx, "Parsing %s", path, params)
			e, err := ParseEpisode(path)
			if err != nil {
				slog.Error(ctx, "Error parsing episode: %s", err, params)
				return nil
			}
			err = i.indexEpisode(e)
			if err != nil {
				slog.Error(ctx, "Error indexing episode: %s", err, params)
				return err
			}
		default:
			// no op
		}
		return nil
	})
	if err != nil {
		slog.Error(ctx, "Error walking library: %s", err, params)
		return nil, err
	}
	return &i, nil
}

func (i *Index) Search(text string) ([]SearchResult, error) {
	i.mu.Lock()
	defer i.mu.Unlock()

	q := bleve.NewMatchQuery(text)
	req := bleve.NewSearchRequest(q)
	res, err := i.index.Search(req)
	if err != nil {
		return nil, err
	}

	sm := make(map[string]SearchResult)
	for _, hit := range res.Hits {
		parts := strings.SplitN(hit.ID, ":", 2)
		eID, ok := i.episodesByCaptionID[parts[0]]
		if !ok {
			return nil, errors.New("not found")
		}

		e := i.episodesByID[eID]
		for _, c := range e.Subtitles {
			if c.ID == parts[0] {
				sm[c.ID] = SearchResult{
					EpisodeID:  eID,
					Caption:    c,
					Confidence: hit.Score,
				}
			}
		}
	}
	sr := make([]SearchResult, 0, len(sm))
	for _, s := range sm {
		sr = append(sr, s)
	}
	return sr, nil
}

func (i *Index) indexEpisode(e domain.Episode) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	i.episodesByID[e.ID] = e

	b := i.index.NewBatch()

	for _, c := range e.Subtitles {
		i.captionsByID[c.ID] = c
		i.episodesByCaptionID[c.ID] = e.ID

		for i, text := range c.Text {
			if text != "" {
				err := b.Index(fmt.Sprintf("%s:%v", c.ID, i), indexRow{
					EpisodeID: e.ID,
					CaptionID: c.ID,
					Text:      strings.ToLower(text),
				})
				if err != nil {
					return err
				}
			}
		}
	}

	return i.index.Batch(b)
}

type indexRow struct {
	EpisodeID string
	CaptionID string
	Text      string
}

func ParseEpisode(path string) (domain.Episode, error) {
	e := domain.Episode{
		ID:   uuid.New().String(),
		Path: path,
	}
	subs, err := subdb.GetSubs(path)
	if err != nil {
		return domain.Episode{}, err
	}
	for _, c := range subs.Captions {
		e.Subtitles = append(e.Subtitles, domain.Caption{
			ID:    uuid.New().String(),
			Start: c.Start,
			End:   c.End,
			Text:  c.Text,
		})
	}
	return e, nil
}
