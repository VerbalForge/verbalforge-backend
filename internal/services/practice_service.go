package services

import (
	"sort"
	"strings"
	"sync"

	"go.mongodb.org/mongo-driver/bson"

	"verbalforge-backend/internal/models"
	"verbalforge-backend/internal/repository"
)

// PracticeService handles practice content operations
type PracticeService struct {
	questionRepo *repository.QuestionRepository
	passageRepo  *repository.PassageRepository
	cacheService *CacheService
}

// NewPracticeService creates a new practice service
func NewPracticeService(
	questionRepo *repository.QuestionRepository,
	passageRepo *repository.PassageRepository,
	cacheService *CacheService,
) *PracticeService {
	return &PracticeService{
		questionRepo: questionRepo,
		passageRepo:  passageRepo,
		cacheService: cacheService,
	}
}

// GetPracticeItems retrieves paginated practice items (questions + passages merged)
func (s *PracticeService) GetPracticeItems(
	page int,
	limit int,
	difficulty string,
	itemType string,
) (*models.PracticeResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Generate cache key
	cacheKey := s.cacheService.GenerateCacheKey(difficulty, itemType)

	// Try to get from cache
	cachedItems, found := s.cacheService.Get(cacheKey)
	var allItems []models.PracticeItem

	if found {
		allItems = cachedItems
	} else {
		// Cache miss - fetch and merge data
		var err error
		allItems, err = s.fetchAndMergeItems(difficulty, itemType)
		if err != nil {
			return nil, err
		}

		// Store in cache
		s.cacheService.Set(cacheKey, allItems)
	}

	// Calculate pagination
	total := int64(len(allItems))
	offset := (page - 1) * limit
	end := offset + limit

	if offset > len(allItems) {
		offset = len(allItems)
	}
	if end > len(allItems) {
		end = len(allItems)
	}

	paginatedItems := allItems[offset:end]
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &models.PracticeResponse{
		Items:      paginatedItems,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
	}, nil
}

// fetchAndMergeItems fetches questions and passages in parallel and merges them
func (s *PracticeService) fetchAndMergeItems(difficulty, itemType string) ([]models.PracticeItem, error) {
	var questions []models.PartialQuestion
	var passages []models.PartialPassage
	var wg sync.WaitGroup
	var questionErr, passageErr error

	// Determine what to fetch based on itemType filter
	fetchQuestions := itemType == "all" || itemType == "" || isQuestionType(itemType)
	fetchPassages := itemType == "all" || itemType == "" || itemType == "reading_comprehension"

	// Fetch questions in parallel
	if fetchQuestions {
		wg.Add(1)
		go func() {
			defer wg.Done()
			questions, questionErr = s.fetchQuestions(difficulty, itemType)
		}()
	}

	// Fetch passages in parallel
	if fetchPassages {
		wg.Add(1)
		go func() {
			defer wg.Done()
			passages, passageErr = s.fetchPassages(difficulty)
		}()
	}

	wg.Wait()

	if questionErr != nil {
		return nil, questionErr
	}
	if passageErr != nil {
		return nil, passageErr
	}

	// Normalize to PracticeItems
	items := make([]models.PracticeItem, 0, len(questions)+len(passages))

	for _, q := range questions {
		items = append(items, models.NewPracticeItemFromQuestion(q))
	}

	for _, p := range passages {
		items = append(items, models.NewPracticeItemFromPassage(p))
	}

	// Sort by created_at descending (newest first)
	sort.Slice(items, func(i, j int) bool {
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})

	return items, nil
}

// fetchQuestions retrieves questions based on filters
func (s *PracticeService) fetchQuestions(difficulty, itemType string) ([]models.PartialQuestion, error) {
	filter := bson.M{}

	if difficulty != "" && difficulty != "all" {
		filter["difficulty_level"] = difficulty
	}

	// Handle question type filters
	if itemType == "text_completion" {
		// Text completion includes single, double, and triple blank types
		filter["question_type"] = bson.M{"$in": []string{
			"text_completion_single",
			"text_completion_double",
			"text_completion_triple",
		}}
	} else if itemType == "sentence_equivalence" {
		filter["question_type"] = "sentence_equivalence"
	} else if itemType == "" || itemType == "all" {
		// CRITICAL: When fetching all questions,
		// exclude passage-based questions (reading comprehension)
		// These questions should only be accessible through their parent passage
		filter["question_type"] = bson.M{"$nin": []string{
			"reading_comprehension_single",
			"reading_comprehension_multiple",
			"reading_comprehension_highlight",
		}}
	}
	// Note: if itemType == "reading_comprehension", we don't fetch any questions
	// (passages will be fetched instead)

	// Fetch all questions (no limit for caching)
	return s.questionRepo.FindPartial(filter, 10000, 0)
}

// fetchPassages retrieves passages based on filters
func (s *PracticeService) fetchPassages(difficulty string) ([]models.PartialPassage, error) {
	filter := bson.M{}

	if difficulty != "" && difficulty != "all" {
		filter["difficulty"] = difficulty
	}

	// Fetch all passages (no limit for caching)
	return s.passageRepo.FindPartial(filter, 10000, 0)
}

// isQuestionType checks if the given type is a valid question type
func isQuestionType(itemType string) bool {
	questionTypes := []string{
		"text_completion",
		"text_completion_single",
		"text_completion_double",
		"text_completion_triple",
		"sentence_equivalence",
	}

	for _, qt := range questionTypes {
		if strings.EqualFold(itemType, qt) {
			return true
		}
	}

	return false
}

// InvalidateCache clears the practice cache for specific filters or all
func (s *PracticeService) InvalidateCache(difficulty, itemType string) {
	if difficulty == "" && itemType == "" {
		s.cacheService.InvalidateAll()
	} else {
		cacheKey := s.cacheService.GenerateCacheKey(difficulty, itemType)
		s.cacheService.Invalidate(cacheKey)
	}
}
