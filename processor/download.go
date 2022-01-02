package processor

import (
	"fmt"
	"os"
	"path"

	"github.com/foamzou/audio-get/consts"
	"github.com/foamzou/audio-get/ffmpeg"
	"github.com/foamzou/audio-get/meta"
	"github.com/foamzou/audio-get/utils"
)

func (p *Processor) download(mediaMeta *meta.MediaMeta) error {
	shouldDownloadAudio, shouldDownloadVideo := p.whichResourceShouldBeDownload(mediaMeta)

	if shouldDownloadAudio {
		_ = p.downloadAudio(mediaMeta)
	}

	if shouldDownloadVideo {
		_ = p.downloadVideo(mediaMeta)
	}
	return nil
}

func (p *Processor) downloadAudio(mediaMeta *meta.MediaMeta) error {
	audioUrl := mediaMeta.Audios[0].Url
	fmt.Printf("-----------------\n start fetch audio: %s\n", audioUrl)
	tempOutputPath, targetOutPath := p.getOutputPaths(mediaMeta.Title, consts.ExtNameMp3)

	if !p.Opts.AddMediaTag && utils.GetExtFromUrl(audioUrl) == consts.ExtNameMp3 {
		err := utils.Wget(audioUrl, targetOutPath, mediaMeta.Headers)
		if err != nil {
			return err
		}
		return nil
	}

	err := utils.Wget(audioUrl, tempOutputPath, mediaMeta.Headers)
	if err != nil {
		return err
	}

	fmt.Printf("start convert, download to : %s\n", tempOutputPath)
	err = ffmpeg.ConvertSingleInput(tempOutputPath, targetOutPath, adjustFileMeta(tempOutputPath, mediaMeta))
	_ = os.Remove(tempOutputPath)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p *Processor) downloadVideo(mediaMeta *meta.MediaMeta) error {
	var tempAudioOutputPath string
	var inputs []string
	if mediaMeta.Videos[0].NeedExtraAudio {
		audioUrl := mediaMeta.Audios[0].Url
		tempAudioOutputPath, _ = p.getOutputPaths(mediaMeta.Title, "mp3")
		err := utils.Wget(audioUrl, tempAudioOutputPath, mediaMeta.Headers)
		if err != nil {
			return err
		}
		inputs = append(inputs, tempAudioOutputPath)
	}

	videoUrl := mediaMeta.Videos[0].Url
	fmt.Printf("-----------------\n start fetch video: %s\n", videoUrl)
	tempOutputPath, targetOutPath := p.getOutputPaths(mediaMeta.Title, consts.ExtNameMp4)

	if !p.Opts.AddMediaTag && utils.GetExtFromUrl(videoUrl) == consts.ExtNameMp4 {
		err := utils.Wget(videoUrl, targetOutPath, mediaMeta.Headers)
		if err != nil {
			return err
		}
		return nil
	}

	err := utils.Wget(videoUrl, tempOutputPath, mediaMeta.Headers)
	if err != nil {
		return err
	}
	inputs = append(inputs, tempOutputPath)

	fmt.Printf("start convert, download to : %s\n", tempOutputPath)
	err = ffmpeg.ConvertMultiInput(inputs, targetOutPath, adjustFileMeta(tempOutputPath, mediaMeta))
	_ = os.Remove(tempOutputPath)
	if tempAudioOutputPath != "" {
		_ = os.Remove(tempAudioOutputPath)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (p *Processor) whichResourceShouldBeDownload(mediaMeta *meta.MediaMeta) (shouldDownloadAudio, shouldDownloadVideo bool) {
	var resourceType string
	if p.Opts.ResourceType == consts.ResourceTypeAuto {
		resourceType = mediaMeta.ResourceType
	} else {
		resourceType = p.Opts.ResourceType
	}
	if resourceType == consts.ResourceTypeAll || resourceType == consts.ResourceTypeAudio {
		shouldDownloadAudio = true
	}
	if resourceType == consts.ResourceTypeAll || resourceType == consts.ResourceTypeVideo {
		shouldDownloadVideo = true
	}
	return
}

func (p *Processor) getOutputPaths(fileTitle string, newExt string) (tempFilePath, targetOutPath string) {
	if p.Opts.Out != "" {
		tempFilePath = p.Opts.Out
	} else {
		// Default: current dir
		dir, err := utils.GetCurrentDir()
		if err != nil {
			panic(err)
		}
		p.Opts.IsDir = true
		tempFilePath = dir
	}

	if p.Opts.IsDir {
		targetOutPath = utils.ModifyFileExt(path.Join(tempFilePath, fileTitle), newExt)
		tempFilePath = path.Join(tempFilePath, fileTitle+"-"+newExt)
	} else {
		targetOutPath = tempFilePath
		tempFilePath = path.Join(path.Dir(tempFilePath), fileTitle+"-"+newExt)
	}

	return
}

func adjustFileMeta(filePath string, mediaMeta *meta.MediaMeta) *ffmpeg.MetaTag {
	metaTag := &ffmpeg.MetaTag{
		Album:  mediaMeta.Album,
		Title:  mediaMeta.Title,
		Artist: mediaMeta.Artist,
	}
	mediaFormat, err := ffmpeg.GetMediaFormat(filePath)

	if err != nil {
		fmt.Println(err)
		return metaTag
	}

	// Priority to use origin meta
	if mediaFormat.Format.Tags.Album != "" {
		metaTag.Album = mediaFormat.Format.Tags.Album
	}
	if mediaFormat.Format.Tags.Artist != "" {
		metaTag.Artist = mediaFormat.Format.Tags.Artist
	}
	if mediaFormat.Format.Tags.Title != "" {
		metaTag.Title = mediaFormat.Format.Tags.Title
	}

	return metaTag
}
