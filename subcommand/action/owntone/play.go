package owntone

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type PlayAction struct {
	name string
	c    *Client
}

// Run

func (a PlayAction) Run(args string) (string, error) {
	if strings.HasPrefix(args, "artist") {
		return a.playRandomArtists()
	} else if strings.HasPrefix(args, "genre") {
		return a.playRandomGenre()
	} else {
		return a.playPlaylist()
	}
}

func (a PlayAction) playRandomGenre() (string, error) {
	msg := []string{"Add"}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	genres, err := a.c.GetGenres()
	if err != nil {
		fmt.Println("error in playRandomGenre")
		return "", err
	}
	index := rand.Intn(len(genres))
	genre := genres[index]
	msg = append(msg, fmt.Sprintf("Genre : %v", genre.Name))
	expression := fmt.Sprintf("genre is \"%s\"", genre.Name)
	err = a.c.AddItem2QueueAndPlay("", expression)
	if err != nil {
		fmt.Println("error in AddItem2QueueAndPlay")
		return "", err
	}
	return strings.Join(msg, " "), nil
}

func (a PlayAction) playRandomArtists() (string, error) {
	msg := []string{"Add"}
	counts, err := a.c.Counts()
	if err != nil {
		return "", err
	}
	if counts.Artists > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		offset := rand.Intn(counts.Artists)
		artist, err := a.c.GetArtist(offset)
		if err != nil {
			fmt.Println("error in playRandomArtists")
			return "", err
		}
		msg = append(msg, fmt.Sprintf("Artist : %v", artist.Name))
		err = a.c.AddItem2QueueAndPlay(artist.Uri, "")
		if err != nil {
			fmt.Println("error in AddItem2QueueAndPlay")
			return "", err
		}
	} else {
		fmt.Println("couldn't get artist")
	}
	return strings.Join(msg, " "), nil
}

func (a PlayAction) playPlaylist() (string, error) {
	// キューに曲がある場合は、そのまま再生
	// キューに曲がない場合は、ランダムにプレイリストを選択してからキューに登録して再生
	status, err := a.c.GetPlayerStatus()
	msg := []string{"Playing music"}
	if err != nil {
		return "", err
	}
	if status.ItemID == 0 {
		//プレイヤーのキューに曲が入っていない状態
		fmt.Print("queue is empty, so playing a randomly selected playlist")
		playlists, err := a.c.GetPlaylists()
		if err != nil {
			fmt.Println("error in playPlaylist")
			return "", err
		}
		if len(playlists) > 0 {
			rand.New(rand.NewSource(time.Now().UnixNano()))
			index := rand.Intn(len(playlists))
			target := playlists[index]
			fmt.Printf("[%v]\n", target.Name)
			msg = append(msg, fmt.Sprintf("from %v.", target.Name))
			err := a.c.AddItem2QueueAndPlay(target.Uri, "")
			if err != nil {
				fmt.Println("error in AddItem2QueueAndPlay")
				return "", err
			}
		} else {
			fmt.Println("playlists is empty")
		}
	} else {
		msg = append(msg, " from queue")
		err = a.c.Play()
		if err != nil {
			fmt.Println("error in Play")
			return "", err
		}
	}
	return strings.Join(msg, " "), nil
}

func NewPlayAction(client *Client) PlayAction {
	return PlayAction{
		name: "Play music on Owntone",
		c:    client,
	}
}
