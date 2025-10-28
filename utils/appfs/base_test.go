package appfs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_Open(t *testing.T) {
	t.Run("file does not exist", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		res, err := appFs.Open("'")

		require.Error(t, err)
		require.True(t, os.IsNotExist(err))
		require.Nil(t, res)
	})

	t.Run("file exists", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		appFs.Fs.Create("/a")

		res, err := appFs.Open("/a")
		require.NoError(t, err)
		require.NotNil(t, res)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ReadDir(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		res, err := appFs.ReadDir("'", false)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path ': open ': file does not exist")
	})

	t.Run("read path error", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		appFs.Fs.Create("/test")
		res, err := appFs.ReadDir("/test", false)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path /test: readdir /test: not a dir")
	})

	t.Run("success", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		appFs.Fs.Create("/a")
		appFs.Fs.Create("/b")
		appFs.Fs.Mkdir("/c", 0755)

		res, err := appFs.ReadDir("/", true)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 2, len(res.Files))
		require.Equal(t, 1, len(res.Directories))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_ReadDirFlat(t *testing.T) {
	t.Run("open error", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		res, err := appFs.ReadDirFlat("'", 1)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to open path ': open ': file does not exist")
	})

	t.Run("read path error", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		appFs.Fs.Create("/test")
		res, err := appFs.ReadDirFlat("/test", 1)

		require.Nil(t, res)
		require.EqualError(t, err, "unable to read path /test: readdir /test: not a dir")
	})

	t.Run("success", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		// Top files/dirs
		appFs.Fs.Mkdir("/1", 0755)
		appFs.Fs.Mkdir("/2/2", 0755)
		appFs.Fs.Mkdir("/3/3/3", 0755)
		appFs.Fs.Mkdir("/4/4/4/4", 0755)
		appFs.Fs.Create("/f1")
		appFs.Fs.Create("/1/f1")
		appFs.Fs.Create("/2/f1")
		appFs.Fs.Create("/2/2/f1")
		appFs.Fs.Create("/3/f1")
		appFs.Fs.Create("/3/3/f1")
		appFs.Fs.Create("/3/3/3/f1")
		appFs.Fs.Create("/4/f1")
		appFs.Fs.Create("/4/4/f1")
		appFs.Fs.Create("/4/4/4/f1")
		appFs.Fs.Create("/4/4/4/4/f1")

		// Depth 0 (same as 1)
		res, err := appFs.ReadDirFlat("/", 0)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 1, len(res))

		// Depth 1
		res, err = appFs.ReadDirFlat("/", 1)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 1, len(res))

		// Depth 10
		res, err = appFs.ReadDirFlat("/", 2)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 5, len(res))

		// Depth 10
		res, err = appFs.ReadDirFlat("/", 10)
		require.NoError(t, err)
		require.NotNil(t, res)
		require.Equal(t, 11, len(res))
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_NonWslDrives(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		// Create WSL directory structure
		paths := []string{}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.NoError(t, err)
		}

		drives, err := appFs.nonWslDrives()

		if errors.Is(err, fmt.Errorf("not implemented yet")) {
			t.Skip("not implemented")
		}

		require.NoError(t, err)
		require.NotEmpty(t, drives)
	})
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func Test_WslDrives(t *testing.T) {
	t.Run("error", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		paths := []string{}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.NoError(t, err)
		}

		drives, err := appFs.wslDrives()

		require.Nil(t, drives)
		require.EqualError(t, err, "unable to open path /mnt: open /mnt: file does not exist")
	})

	t.Run("success", func(t *testing.T) {
		appFs := New(afero.NewMemMapFs())

		// Create WSL directory structure
		paths := []string{"/mnt/c", "/mnt/d", "/mnt/wsl", "/mnt/wslg"}
		for _, p := range paths {
			err := appFs.Fs.MkdirAll(p, os.ModePerm)
			require.NoError(t, err)
		}

		drives, err := appFs.wslDrives()
		require.NoError(t, err)
		require.Len(t, drives, 3)
		require.ElementsMatch(t, []string{"/", filepath.Join("/mnt", "c"), filepath.Join("/mnt", "d")}, drives)
	})
}
