package pkg

import (
	"context"
	"fmt"

	"github.com/zmb3/spotify/v2"
)

type Summary struct {
	TopGenres  []string `json:"top_genres"`
	TopArtists []string `json:"top_artists"`
	Archetype  string
}

func AnalyzeUser(ctx context.Context, client *spotify.Client) (*Summary, error) {
	topPage, errTop := client.CurrentUsersTopTracks(ctx, spotify.Limit(20))
	if errTop != nil {
		return nil, fmt.Errorf("could not fetch first page of top tracks: %w", errTop)
	}

	var topArtists, topGenres []string

	top, err := AnalyzeTop(ctx, client, topPage)
	if err == nil && len(top.TopArtists) > 0 {
		topArtists = top.TopArtists
		topGenres = top.TopGenres
	} else {
		savedPage, errSaved := client.CurrentUsersTracks(ctx, spotify.Limit(5))
		if errSaved != nil {
			return nil, fmt.Errorf("could not fetch first page of top tracks: %w", errSaved)
		}

		saved, err2 := AnalyzeSaved(ctx, client, savedPage)
		if err2 != nil {
			return nil, fmt.Errorf("could not analyze top tracks (%v) or saved tracks (%w)", err, err2)
		}
		topArtists = saved.TopArtists
		topGenres = saved.TopGenres
	}

	archetype := AnalyzeArchetype(topGenres)

	return &Summary{
		TopGenres:  topGenres,
		TopArtists: topArtists,
		Archetype:  archetype,
	}, nil
}
