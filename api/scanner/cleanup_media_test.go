package scanner_test

import (
	"os"
	"path"
	"testing"

	"github.com/otiai10/copy"
	"github.com/photoview/photoview/api/graphql/models"
	"github.com/photoview/photoview/api/test_utils"
	"github.com/stretchr/testify/assert"
)

func TestCleanupMedia(t *testing.T) {
	test_utils.FilesystemTest(t)
	db := test_utils.DatabaseTest(t)

	test_dir := t.TempDir()
	copy.Copy("./test_data", test_dir)

	countAllMedia := func() int {
		var all_media []*models.Media
		if !assert.NoError(t, db.Find(&all_media).Error) {
			return -1
		}
		return len(all_media)
	}

	countAllMediaURLs := func() int {
		var all_media_urls []*models.MediaURL
		if !assert.NoError(t, db.Find(&all_media_urls).Error) {
			return -1
		}
		return len(all_media_urls)
	}

	pass := "1234"
	user1, err := models.RegisterUser(db, "user1", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	user2, err := models.RegisterUser(db, "user2", &pass, true)
	if !assert.NoError(t, err) {
		return
	}

	root_album := models.Album{
		Title: "root album",
		Path:  test_dir,
	}

	if !assert.NoError(t, db.Save(&root_album).Error) {
		return
	}

	err = db.Model(user1).Association("Albums").Append(&root_album)
	if !assert.NoError(t, err) {
		return
	}
	err = db.Model(user2).Association("Albums").Append(&root_album)
	if !assert.NoError(t, err) {
		return
	}

	t.Run("Modify albums", func(t *testing.T) {

		test_utils.RunScannerOnUser(t, db, user1)
		assert.Equal(t, 9, countAllMedia())
		assert.Equal(t, 18, countAllMediaURLs())

		// move faces directory
		assert.NoError(t, os.Rename(path.Join(test_dir, "faces"), path.Join(test_dir, "faces_moved")))
		test_utils.RunScannerAll(t, db)
		assert.Equal(t, 9, countAllMedia())
		assert.Equal(t, 18, countAllMediaURLs())

		// remove faces_moved directory
		assert.NoError(t, os.RemoveAll(path.Join(test_dir, "faces_moved")))
		test_utils.RunScannerAll(t, db)
		assert.Equal(t, 3, countAllMedia())
		assert.Equal(t, 6, countAllMediaURLs())
	})

	// t.Run("Modify images", func(t *testing.T) {

	// })
}
