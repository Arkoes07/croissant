package quiz

import (
	"errors"
	"math/rand"
	"time"

	"github.com/Arkoes07/croissant/internal/song"
)

// below are the known Generator errors
var (
	ErrPoolTooSmall = errors.New("song pool is too small to generate questions with enough distractors")
)

// GeneratorConfig holds tunable values for question generation.
type GeneratorConfig struct {
	// QuestionCount is the number of questions per quiz.
	QuestionCount int
	// ChoiceCount is the number of choices per question (including the correct one).
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
	if cfg.ChoiceCount > cfg.QuestionCount+1 {
		// Need at least ChoiceCount songs in the pool.
	}

	return &Generator{
		cfg: cfg,
		rng: rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

// Generate builds a slice of Questions from the given song pool.
// The pool must contain at least QuestionCount + ChoiceCount - 1 songs so
// each question can have unique distractors not used as the correct answer
// in other questions.
func (g *Generator) Generate(pool []song.Song) ([]Question, error) {
	required := g.cfg.QuestionCount + g.cfg.ChoiceCount - 1
	if len(pool) < required {
		return nil, ErrPoolTooSmall
	}

	// Shuffle a copy so we don't mutate the caller's slice.
	shuffled := make([]song.Song, len(pool))
	copy(shuffled, pool)
	g.rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})

	// The first QuestionCount songs are the correct answers, one per question.
	// The remaining songs form the distractor pool.
	correctSongs := shuffled[:g.cfg.QuestionCount]
	distractorPool := shuffled[g.cfg.QuestionCount:]

	questions := make([]Question, g.cfg.QuestionCount)
	for i, correct := range correctSongs {
		// Pick ChoiceCount-1 distractors for this question.
		distractorCount := g.cfg.ChoiceCount - 1
		distractors := make([]song.Song, distractorCount)
		copy(distractors, distractorPool[i*distractorCount:(i+1)*distractorCount])

		// Build the choices slice and insert the correct song at a random position.
		correctIdx := g.rng.Intn(g.cfg.ChoiceCount)
		choices := make([]song.Song, g.cfg.ChoiceCount)
		di := 0
		for ci := range choices {
			if ci == correctIdx {
				choices[ci] = correct
			} else {
				choices[ci] = distractors[di]
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
