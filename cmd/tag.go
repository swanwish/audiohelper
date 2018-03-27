package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/bogem/id3v2"
	"github.com/swanwish/go-common/logs"
	"github.com/swanwish/go-common/utils"
	"github.com/urfave/cli"
)

var (
	ListTagsCmd = cli.Command{
		Name:        "list-tags",
		Usage:       "List audio file tags",
		Description: "This command will search audio files in the specified path, and list the tags of the audio files, the index parameter: -1 means the parent dir of the audio file, -2 means parent of the parent dir",
		Action:      listTags,
		Flags: []cli.Flag{
			stringFlag("path", "", "The path to check"),
		},
	}
	SetTagsCmd = cli.Command{
		Name:        "set-tags",
		Usage:       "Set the tag for audio files",
		Description: "This command will set the tags for the audio files",
		Action:      setTags,
		Flags: []cli.Flag{
			stringFlag("path", "", "The path to check"),
			stringFlag("album", "", "The album of the audio file"),
			stringFlag("year", "", "The year of the audio file"),
			stringFlag("artist", "", "The artist of the audio file"),
			intFlag("albumIndex", 0, "The index of the album name"),
			intFlag("yearIndex", 0, "The index of the album year"),
			intFlag("artistIndex", 0, "The index of the artist"),
		},
	}
)

func listTags(c *cli.Context) error {
	dirPath := c.String("path")
	if dirPath == "" {
		dirPath = "."
	}
	if !utils.FileExists(dirPath) {
		message := fmt.Sprintf("The path %s does not exists", dirPath)
		logs.Errorf(message)
		return cli.NewExitError(message, 1)
	}
	err := listTagsAt(dirPath)
	if err != nil {
		message := fmt.Sprintf("Failed to list tags at path %s, the error is %#v", dirPath, err)
		logs.Errorf(message)
		return cli.NewExitError(message, 1)
	}
	return nil
}

func listTagsAt(dirPath string) error {
	list, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logs.Errorf("Failed to list dir, the error is %v", err)
		return err
	}
	for _, item := range list {
		itemPath := filepath.Join(dirPath, item.Name())
		logs.Debugf("Find item %s", item.Name())
		if item.IsDir() {
			if err = listTagsAt(itemPath); err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(item.Name(), ".mp3") {
				tag, err := id3v2.Open(itemPath, id3v2.Options{Parse: true})
				if err != nil {
					logs.Errorf("Error while opening mp3 file: %s, the error is %#v", itemPath, err)
					return err
				}
				fmt.Printf("item: %s\nalbum: %s\nyear: %s\nartist: %s\ntitle: %s\n", item.Name(), tag.Album(), tag.Year(), tag.Artist(), tag.Title())
				tag.Close()
			}
		}
	}
	return nil
}

func setTags(c *cli.Context) error {
	dirPath := c.String("path")
	if dirPath == "" {
		dirPath = "."
	}
	if !utils.FileExists(dirPath) {
		message := fmt.Sprintf("The path %s does not exists", dirPath)
		logs.Errorf(message)
		return cli.NewExitError(message, 1)
	}
	err := setTagsAt(dirPath, c)
	if err != nil {
		message := fmt.Sprintf("Failed to set tags at path %s, the error is %#v", dirPath, err)
		logs.Errorf(message)
		return cli.NewExitError(message, 1)
	}
	return nil
}

func setTagsAt(dirPath string, c *cli.Context) error {
	album := c.String("album")
	year := c.String("year")
	artist := c.String("artist")
	albumIndex := c.Int64("albumIndex")
	yearIndex := c.Int64("yearIndex")
	artistIndex := c.Int64("artistIndex")

	list, err := ioutil.ReadDir(dirPath)
	if err != nil {
		logs.Errorf("Failed to list dir, the error is %v", err)
		return err
	}

	for _, item := range list {
		itemPath := filepath.Join(dirPath, item.Name())
		logs.Debugf("Find item %s", item.Name())
		if item.IsDir() {
			if err = setTagsAt(itemPath, c); err != nil {
				return err
			}
		} else {
			if strings.HasSuffix(item.Name(), ".mp3") {
				tag, err := id3v2.Open(itemPath, id3v2.Options{Parse: true})
				if err != nil {
					logs.Errorf("Error while opening mp3 file: %s, the error is %#v", itemPath, err)
					return err
				}
				title := strings.TrimRight(filepath.Base(item.Name()), filepath.Ext(item.Name()))
				tag.SetTitle(title)
				parts := strings.Split(dirPath, "/")
				itemAlbum := getPathItem(parts, albumIndex, album)
				itemArtist := getPathItem(parts, artistIndex, artist)
				itemYear := getPathItem(parts, yearIndex, year)
				if itemAlbum != "" {
					tag.SetAlbum(itemAlbum)
				}
				if itemArtist != "" {
					tag.SetArtist(itemArtist)
				}
				if itemYear != "" {
					tag.SetYear(itemYear)
				}
				if err = tag.Save(); err != nil {
					logs.Errorf("Failed to save tag (title: %s, album: %s, artist: %s, year: %s) on item %s, the error is %#v", title, itemAlbum, itemArtist, itemYear, itemPath, err)
					tag.Close()
					continue
				}
				tag.Close()
			}
		}
	}
	return nil
}

func getPathItem(pathItems []string, itemIndex int64, defaultValue string) string {
	if defaultValue != "" {
		return defaultValue
	}
	if itemIndex < 0 {
		index := int64(len(pathItems)) + itemIndex
		if index >= 0 {
			return pathItems[index]
		}
	}
	return ""
}
