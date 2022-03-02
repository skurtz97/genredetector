package client

import "sort"

type itemsSorter struct {
	items []Artist
	by    func(i1, i2 *Artist) bool // closure used in less method
}

type By func(i1, i2 *Artist) bool

// part of sort interface
func (s *itemsSorter) Len() int {
	return len(s.items)
}

// part of sort interface
func (s *itemsSorter) Swap(i, j int) {
	s.items[i], s.items[j] = s.items[j], s.items[i]
}

// part of sort interface. implemented by calling "by" closure in sorter, we define this inline where we need to sort
func (s *itemsSorter) Less(i, j int) bool {
	return s.by(&s.items[i], &s.items[j])
}

func (by By) Sort(items []Artist) {
	is := &itemsSorter{
		items: items,
		by:    by,
	}
	sort.Sort(is)
}

// sorts artists according to popularity, with highest popularity at the top/beginning of the returned slice
func SortArtists(artists []Artist) []Artist {
	popularityDescending := func(i1, i2 *Artist) bool {
		return i1.Popularity > i2.Popularity
	}

	By(popularityDescending).Sort(artists)

	return artists
}
