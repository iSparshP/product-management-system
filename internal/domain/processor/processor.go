package processor

import "github.com/iSparshP/product-management-system/internal/domain/model"

type ImageProcessor interface {
	ProcessImageTask(task model.ImageProcessingTask) error
}
