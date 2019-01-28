package insta

import (
	"errors"
	"log"

	"github.com/ahmdrz/goinsta"
)

// Feed represents a user feed
type Feed struct {
	ig              *goinsta.Instagram
	ID              int64
	mainNextMaxID   string
	mainDone        bool
	taggedNextMaxID string
	taggedDone      bool
}

// MainNext returns expanded (no carousel) medi items
func (f *Feed) MainNext() ([]Media, error) {
	if f.mainDone {
		return []Media{}, errors.New("No more media")
	}
	res, err := f.ig.UserFeed(f.ID, f.mainNextMaxID, "")
	if err != nil {
		return []Media{}, err
	}

	f.mainNextMaxID = res.NextMaxID
	if f.mainNextMaxID == "" {
		f.mainDone = true
	}

	items := []Media{}
	for _, item := range res.Items {
		switch item.MediaType {
		case 1:
			// Photo
			mw := 0
			murl := ""
			for _, iv := range item.ImageVersions2.Candidates {
				if iv.Width > mw {
					mw = iv.Width
					murl = iv.URL
				}
			}

			items = append(items, Media{
				ID:   item.ID,
				Type: item.MediaType,
				URL:  murl,
			})
		case 2:
			// Video
			mw := 0
			murl := ""
			for _, vv := range item.VideoVersions {
				if vv.Width > mw {
					mw = vv.Width
					murl = vv.URL
				}
			}
			items = append(items, Media{
				ID:   item.ID,
				Type: item.MediaType,
				URL:  murl,
			})
		case 8:
			// carousel
			for _, ci := range item.CarouselMedia {
				switch ci.MediaType {
				case 1:
					// Photo
					mw := 0
					murl := ""
					for _, iv := range ci.ImageVersions.Candidates {
						if iv.Width > mw {
							mw = iv.Width
							murl = iv.URL
						}
					}

					items = append(items, Media{
						ID:   ci.ID,
						Type: ci.MediaType,
						URL:  murl,
					})
				case 2:
					mw := 0
					murl := ""
					for _, vv := range ci.VideoVersions {
						if vv.Width > mw {
							mw = vv.Width
							murl = vv.URL
						}
					}
					items = append(items, Media{
						ID:   ci.ID,
						Type: ci.MediaType,
						URL:  murl,
					})
				default:
					log.Println("Unkown media type")

				}
			}
		default:
			log.Println("Unkown media type")
		}
	}
	return items, nil
}

// TaggedNext returns expanded (no carousel) medi items
func (f *Feed) TaggedNext() ([]Media, error) {
	if f.taggedDone {
		return []Media{}, errors.New("No more media")
	}
	res, err := f.ig.UserTaggedFeed(f.ID, f.taggedNextMaxID, "")
	if err != nil {
		return []Media{}, err
	}

	f.taggedNextMaxID = res.NextMaxID
	if f.taggedNextMaxID == "" {
		f.taggedDone = true
	}

	items := []Media{}
	for _, item := range res.Items {
		switch item.MediaType {
		case 1:
			// Photo
			mw := 0
			murl := ""
			for _, iv := range item.ImageVersions2.Candidates {
				if iv.Width > mw {
					mw = iv.Width
					murl = iv.URL
				}
			}

			items = append(items, Media{
				ID:   item.ID,
				Type: item.MediaType,
				URL:  murl,
			})
		case 2:
			// Video
			mw := 0
			murl := ""
			for _, vv := range item.VideoVersions {
				if vv.Width > mw {
					mw = vv.Width
					murl = vv.URL
				}
			}
			items = append(items, Media{
				ID:   item.ID,
				Type: item.MediaType,
				URL:  murl,
			})
		case 8:
			// carousel
			for _, ci := range item.CarouselMedia {
				switch ci.MediaType {
				case 1:
					// Photo
					mw := 0
					murl := ""
					for _, iv := range ci.ImageVersions.Candidates {
						if iv.Width > mw {
							mw = iv.Width
							murl = iv.URL
						}
					}

					items = append(items, Media{
						ID:   ci.ID,
						Type: ci.MediaType,
						URL:  murl,
					})
				case 2:
					mw := 0
					murl := ""
					for _, vv := range ci.VideoVersions {
						if vv.Width > mw {
							mw = vv.Width
							murl = vv.URL
						}
					}
					items = append(items, Media{
						ID:   ci.ID,
						Type: ci.MediaType,
						URL:  murl,
					})
				default:
					log.Println("Unkown media type")

				}
			}
		default:
			log.Println("Unkown media type")
		}
	}
	return items, nil
}
