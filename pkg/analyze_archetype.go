package pkg

var RISKLORD = []string{"metal", "hardcore", "punk", "drum and bass"}
var STRATEGIST = []string{"techno", "classical", "ambient", "progressive"}
var OPTIMIST = []string{"funk", "disco", "pop", "groove"}
var HUSTLER = []string{"rap", "trap", "drill", "afrobeat"}
var ZEN_INVESTOR = []string{"jazz", "lo-fi", "indie folk", "instrumental"}
var CHAOS_SELECTOR = []string{"hyperpop", "experimental", "electronic", "indie alt"}
var OUTLIER = []string{}

func AnalyzeArchetype(topGenres []string) string {
	genreSet := make(map[string]struct{}, len(topGenres))
	for _, g := range topGenres {
		genreSet[g] = struct{}{}
	}

	archetypes := map[string][]string{
		"RISKLORD":       RISKLORD,
		"STRATEGIST":     STRATEGIST,
		"OPTIMIST":       OPTIMIST,
		"HUSTLER":        HUSTLER,
		"ZEN_INVESTOR":   ZEN_INVESTOR,
		"CHAOS_SELECTOR": CHAOS_SELECTOR,
		"OUTLIER":        OUTLIER,
	}

	counts := make(map[string]int, len(archetypes))
	maxMatches := 0

	for name, list := range archetypes {
		count := 0
		for _, g := range list {
			if _, ok := genreSet[g]; ok {
				count++
			}
		}

		counts[name] = count
		if count > maxMatches {
			maxMatches = count
		}
	}

	var best []string
	for name, cnt := range counts {
		if cnt == maxMatches {
			best = append(best, name)
		}
	}

	if maxMatches == 0 || len(best) != 1 {
		return "OUTLIER"
	}

	return best[0]
}
