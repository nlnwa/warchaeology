package filewalker2

import (
	"io/fs"

	"github.com/spf13/afero"
)

type FileToProcess struct {
	Path     string
	FileInfo fs.FileInfo
	Error    error
}

func PopulateChannelWithFilesToProcess(fileSystem afero.Fs, dirName string, filePathChannel chan FileToProcess) error {
	err := afero.Walk(
		fileSystem, dirName, func(path string, fileInfo fs.FileInfo, err error) error {
			filePathChannel <- FileToProcess{
				path,
				fileInfo,
				err,
			}
			return nil
		},
	)
	close(filePathChannel)
	return err
}
