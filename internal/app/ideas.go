package app

type IdeaStorage struct {
	ideas              []string
	ranks              map[int][]int
	currentRankingIdea int
}

func NewIdeaStorage() IdeaStorage {
	return IdeaStorage{
		ideas: make([]string, 0),
		ranks: map[int][]int{1: make([]int, 0), 2: make([]int, 0)},
	}
}

func (is IdeaStorage) AreAllIdeasRanked() bool {
	return is.currentRankingIdea >= len(is.ideas)
}

func (is *IdeaStorage) RankCurrentIdea(isLiked bool) {
	if isLiked {
		is.ranks[1] = append(is.ranks[1], is.currentRankingIdea)
	} else {
		is.ranks[2] = append(is.ranks[2], is.currentRankingIdea)
	}

	is.currentRankingIdea++
}

func (is *IdeaStorage) Add(idea string) {
	is.ideas = append(is.ideas, idea)
}

func (is IdeaStorage) GetGoodIdeas() *[]string {
	goodIdeas := make([]string, 0, 2)
	for _, v := range is.ranks[1] {
		goodIdeas = append(goodIdeas, is.ideas[v])
	}
	return &goodIdeas
}

func (is IdeaStorage) GetBadIdeas() *[]string {
	goodIdeas := make([]string, 0, 2)
	for _, v := range is.ranks[2] {
		goodIdeas = append(goodIdeas, is.ideas[v])
	}
	return &goodIdeas
}
