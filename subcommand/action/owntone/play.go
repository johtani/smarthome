package owntone

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
)

type PlayAction struct {
	name string
	c    *Client
}

// Run

func (a PlayAction) Run(ctx context.Context, args string) (string, error) {
	ctx, span := otel.Tracer("owntone").Start(ctx, "PlayAction.Run")
	defer span.End()
	switch {
	case strings.HasPrefix(args, "artist"):
		return a.playRandomArtists(ctx)
	case strings.HasPrefix(args, "genre"):
		return a.playRandomGenre(ctx)
	default:
		return a.playPlaylist(ctx)
	}
}

func (a PlayAction) playRandomGenre(ctx context.Context) (string, error) {
	msg := []string{"Add"}
	rand.New(rand.NewSource(time.Now().UnixNano()))
	genres, err := a.c.GetGenres(ctx)
	if err != nil {
		return "", fmt.Errorf("error in playRandomGenre\n %v", err)
	}
	index := rand.Intn(len(genres))
	genre := genres[index]
	msg = append(msg, fmt.Sprintf("Genre : %v", genre.Name))
	expression := fmt.Sprintf("genre is \"%s\"", genre.Name)
	err = a.c.AddItem2QueueAndPlay(ctx, "", expression)
	if err != nil {
		return "", fmt.Errorf("error in AddItem2QueueAndPlay\n %v", err)
	}
	return strings.Join(msg, " "), nil
}

func (a PlayAction) playRandomArtists(ctx context.Context) (string, error) {
	msg := []string{"Add"}
	counts, err := a.c.Counts(ctx)
	if err != nil {
		return "", err
	}
	if counts.Artists > 0 {
		rand.New(rand.NewSource(time.Now().UnixNano()))
		offset := rand.Intn(counts.Artists)
		artist, err := a.c.GetArtist(ctx, offset)
		if err != nil {
			return "", fmt.Errorf("error in playRandomArtists\n %v", err)
		}
		msg = append(msg, fmt.Sprintf("Artist : %v", artist.Name))
		err = a.c.AddItem2QueueAndPlay(ctx, artist.URI, "")
		if err != nil {
			return "", fmt.Errorf("error in AddItem2QueueAndPlay\n %v", err)
		}
	} else {
		msg = append(msg, "couldn't get artist")
	}
	return strings.Join(msg, " "), nil
}

func (a PlayAction) playPlaylist(ctx context.Context) (string, error) {
	// キューに曲がある場合は、そのまま再生
	// キューに曲がない場合は、ランダムにプレイリストを選択してからキューに登録して再生
	status, err := a.c.GetPlayerStatus(ctx)
	msg := []string{"Playing music"}
	if err != nil {
		return "", err
	}
	if status.ItemID == 0 {
		// プレイヤーのキューに曲が入っていない状態
		// fmt.Print("queue is empty, so playing a randomly selected playlist")
		playlists, err := a.c.GetPlaylists(ctx)
		if err != nil {
			return "", fmt.Errorf("error in playPlaylist\n %v", err)
		}
		if len(playlists) > 0 {
			rand.New(rand.NewSource(time.Now().UnixNano()))
			index := rand.Intn(len(playlists))
			target := playlists[index]
			msg = append(msg, "from "+target.Name+".")
			err := a.c.AddItem2QueueAndPlay(ctx, target.URI, "")
			if err != nil {
				return "", fmt.Errorf("error in AddItem2QueueAndPlay(%v)\n %v", target.Name, err)
			}
		} else {
			msg = append(msg, "playlists is empty")
		}
	} else {
		msg = append(msg, " from queue")
		err = a.c.Play(ctx)
		if err != nil {
			return "", fmt.Errorf("error in Play\n %v", err)
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
