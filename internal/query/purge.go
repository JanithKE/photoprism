package query

import (
	"time"

	"github.com/photoprism/photoprism/internal/entity"
)

// PurgeOrphans removes orphan database entries.
func PurgeOrphans() error {
	// Remove files without a photo.
	start := time.Now()
	if count, err := PurgeOrphanFiles(); err != nil {
		return err
	} else if count > 0 {
		log.Warnf("purge: removed %d orphan files [%s]", count, time.Since(start))
	} else {
		log.Infof("purge: no orphan files [%s]", time.Since(start))
	}

	// Remove duplicates without an original file.
	if err := PurgeOrphanDuplicates(); err != nil {
		return err
	}

	// Remove unused countries.
	if err := PurgeOrphanCountries(); err != nil {
		return err
	}

	// Remove unused cameras.
	if err := PurgeOrphanCameras(); err != nil {
		return err
	}

	// Remove unused camera lenses.
	if err := PurgeOrphanLenses(); err != nil {
		return err
	}

	return nil
}

// PurgeOrphanFiles removes files without a photo from the index.
func PurgeOrphanFiles() (count int, err error) {
	files, err := OrphanFiles()

	if err != nil {
		return count, err
	}

	for i := range files {
		if err = files[i].DeletePermanently(); err != nil {
			return count, err
		}

		count++
	}

	return count, err
}

// PurgeOrphanDuplicates deletes all files from the duplicates table that don't exist in the files table.
func PurgeOrphanDuplicates() error {
	return UnscopedDb().Delete(
		entity.Duplicate{},
		"file_hash NOT IN (SELECT file_hash FROM files WHERE file_missing = 0 AND deleted_at IS NULL)").Error
}

// PurgeOrphanCountries removes countries without any photos.
func PurgeOrphanCountries() error {
	entity.FlushCountryCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM countries WHERE country_slug <> ? AND id NOT IN (SELECT photo_country FROM photos)`, entity.UnknownCountry.CountrySlug).Error
	}
}

// PurgeOrphanCameras removes cameras without any photos.
func PurgeOrphanCameras() error {
	entity.FlushCameraCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM cameras WHERE camera_slug <> ? AND id NOT IN (SELECT camera_id FROM photos)`, entity.UnknownCamera.CameraSlug).Error
	}
}

// PurgeOrphanLenses removes cameras without any photos.
func PurgeOrphanLenses() error {
	entity.FlushLensCache()
	switch DbDialect() {
	default:
		return UnscopedDb().Exec(`DELETE FROM lenses WHERE lens_slug <> ? AND id NOT IN (SELECT lens_id FROM photos)`, entity.UnknownLens.LensSlug).Error
	}
}
