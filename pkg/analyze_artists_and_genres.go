package pkg

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/zmb3/spotify/v2"
)

type Top struct {
	TopGenres  []string `json:"top_genres"`
	TopArtists []string `json:"top_artists"`
}

func AnalyzeTop(ctx context.Context, client *spotify.Client, topPage *spotify.FullTrackPage) (*Top, error) {
	allTracks := topPage.Tracks

	for {
		allTracks = append(allTracks, topPage.Tracks...)

		if topPage.Next == "" {
			break
		}

		if err := client.NextPage(ctx, topPage); err != nil {
			return nil, fmt.Errorf("could not fetch next page of top tracks: %w", err)
		}
	}

	if len(allTracks) == 0 {
		return nil, errors.New("no top tracks found")
	}

	trackIDs := make([]spotify.ID, 0, len(allTracks))
	artistIDs := make([]spotify.ID, 0, len(allTracks))
	for _, t := range allTracks {
		trackIDs = append(trackIDs, t.ID)
		if len(t.Artists) > 0 {
			artistIDs = append(artistIDs, t.Artists[0].ID)
		}
	}

	artists, err := batchFetchArtists(ctx, client, artistIDs, 20)
	if err != nil {
		return nil, fmt.Errorf("could not fetch artist data: %w", err)
	}

	topGenres := GenresCount(artists)

	artistCount := make(map[string]int, len(allTracks))
	for _, t := range allTracks {
		if len(t.Artists) > 0 {
			artistCount[t.Artists[0].Name]++
		}
	}
	topArtists := make([]string, 0, len(artistCount))
	for name := range artistCount {
		topArtists = append(topArtists, name)
	}
	sort.Slice(topArtists, func(i, j int) bool {
		return artistCount[topArtists[i]] > artistCount[topArtists[j]]
	})
	if len(topArtists) > 5 {
		topArtists = topArtists[:5]
	}

	return &Top{
		TopGenres:  topGenres,
		TopArtists: topArtists,
	}, nil
}

func AnalyzeSaved(ctx context.Context, client *spotify.Client, savedPage *spotify.SavedTrackPage) (*Top, error) {
	allTracks := savedPage.Tracks

	for {
		allTracks = append(allTracks, savedPage.Tracks...)

		if savedPage.Next == "" {
			break
		}

		if err := client.NextPage(ctx, savedPage); err != nil {
			return nil, fmt.Errorf("could not fetch next page of top tracks: %w", err)
		}
	}

	if len(allTracks) == 0 {
		return nil, errors.New("no top tracks found")
	}

	trackIDs := make([]spotify.ID, 0, len(allTracks))
	artistIDs := make([]spotify.ID, 0, len(allTracks))
	for _, t := range allTracks {
		trackIDs = append(trackIDs, t.ID)
		if len(t.Artists) > 0 {
			artistIDs = append(artistIDs, t.Artists[0].ID)
		}
	}

	artists, err := batchFetchArtists(ctx, client, artistIDs, 20)
	if err != nil {
		return nil, fmt.Errorf("could not fetch artist data: %w", err)
	}

	topGenres := GenresCount(artists)

	artistCount := make(map[string]int, len(allTracks))
	for _, t := range allTracks {
		if len(t.Artists) > 0 {
			artistCount[t.Artists[0].Name]++
		}
	}
	topArtists := make([]string, 0, len(artistCount))
	for name := range artistCount {
		topArtists = append(topArtists, name)
	}
	sort.Slice(topArtists, func(i, j int) bool {
		return artistCount[topArtists[i]] > artistCount[topArtists[j]]
	})
	if len(topArtists) > 5 {
		topArtists = topArtists[:5]
	}

	return &Top{
		TopGenres:  topGenres,
		TopArtists: topArtists,
	}, nil
}

func batchFetchArtists(ctx context.Context, client *spotify.Client, ids []spotify.ID, maxPerCall int) ([]*spotify.FullArtist, error) {
	var all []*spotify.FullArtist
	for start := 0; start < len(ids); start += maxPerCall {
		end := start + maxPerCall
		if end > len(ids) {
			end = len(ids)
		}
		chunk := ids[start:end]
		arts, err := client.GetArtists(ctx, chunk...)
		if err != nil {
			return nil, fmt.Errorf("chunk %d..%d: %w, bearer %s", start, end, err, client.Token)
		}
		all = append(all, arts...)
	}
	return all, nil
}

func GenresCount(artists []*spotify.FullArtist) []string {
	genreCount := map[string]int{}
	for _, art := range artists {
		for _, g := range art.Genres {
			genreCount[g]++
		}
	}
	type genreCounts struct {
		Genre string
		Count int
	}
	var sorted []genreCounts
	for g, c := range genreCount {
		sorted = append(sorted, genreCounts{Genre: g, Count: c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Count > sorted[j].Count
	})
	topGenres := []string{}
	for i := 0; i < len(sorted) && i < 5; i++ {
		topGenres = append(topGenres, sorted[i].Genre)
	}

	return topGenres
}
