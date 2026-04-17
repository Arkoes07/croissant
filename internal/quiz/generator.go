package quiz

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Arkoes07/croissant/internal/song"
)

// below are the known Generator errors
var (
	// ErrPoolTooSmall is returned when the song pool has fewer songs than
	// max(QuestionCount, ChoiceCount), which is the minimum needed to generate
	// a valid quiz.
	ErrPoolTooSmall = errors.New("song pool is too small to generate questions with enough distractors")
)

// GeneratorConfig holds tunable values for question generation.
type GeneratorConfig struct {
	// QuestionCount is the number of questions per quiz.
	QuestionCount int
	// ChoiceCount is the number of answer choices per question (including the correct one).
	ChoiceCount int
}

// Generator builds quiz questions from a pool of songs.
type Generator struct {
	cfg GeneratorConfig
	rng *rand.Rand
}

// NewGenerator creates a Generator with the given config.
func NewGenerator(cfg GeneratorConfig) (*Generator, error) {
	if cfg.QuestionCount <= 0 {
		cfg.QuestionCount = 10
	}
	if cfg.ChoiceCount <= 1 {
		cfg.ChoiceCount = 4
	}

	return &Generator{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Generate builds a slice of Questions from the given song pool.
//
// The pool must contain at least max(QuestionCount, ChoiceCount) songs.
// Distractors are picked randomly per question from the full pool (excluding
// the correct song), so they may repeat across questions — that is intentional
// when the pool is small.
func (g *Generator) Generate(pool []song.Song) ([]Question, error) {
	minRequired := g.cfg.QuestionCount
	if g.cfg.ChoiceCount > minRequired {
		minRequired = g.cfg.ChoiceCount
	}
	if len(pool) < minRequired {
		return nil, ErrPoolTooSmall
	}

	// Shuffle a copy so we don't mutate the caller's slice.
	shuffled := make([]song.Song, len(pool))
	copy(shuffled, pool)
	g.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// The first QuestionCount songs are the correct answers, one per question.
	correctSongs := shuffled[:g.cfg.QuestionCount]

	questions := make([]Question, g.cfg.QuestionCount)
	for i, correct := range correctSongs {
		// Build a distractor pool from all songs except the current correct one.
		distractors := make([]song.Song, 0, len(shuffled)-1)
		for _, s := range shuffled {
			if s.Title != correct.Title {
				distractors = append(distractors, s)
			}
		}

		// Shuffle distractors and take the first ChoiceCount-1.
		g.rng.Shuffle(len(distractors), func(a, b int) {
			distractors[a], distractors[b] = distractors[b], distractors[a]
		})
		chosen := distractors[:g.cfg.ChoiceCount-1]

		// Place the correct song at a random position within the choices.
		correctIdx := g.rng.Intn(g.cfg.ChoiceCount)
		choices := make([]song.Song, g.cfg.ChoiceCount)
		di := 0
		for ci := range choices {
			if ci == correctIdx {
				choices[ci] = correct
			} else {
				choices[ci] = chosen[di]
				di++
			}
		}

		questions[i] = Question{
			Choices:    choices,
			CorrectIdx: correctIdx,
			AudioURL:   correct.PreviewURL,
		}
	}

	return questions, nil
}
