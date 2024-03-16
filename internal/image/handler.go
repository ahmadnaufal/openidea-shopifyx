package image

import (
	"strings"

	"github.com/ahmadnaufal/openidea-shopifyx/internal/model"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/jwt"
	"github.com/ahmadnaufal/openidea-shopifyx/pkg/s3"
	"github.com/gofiber/fiber/v2"
)

var (
	S3ProviderImpl *s3.S3Provider
)

func RegisterRoute(r *fiber.App, jwtProvider jwt.JWTProvider) {
	imageGroup := r.Group("/v1/image")

	authMiddleware := jwtProvider.Middleware()
	imageGroup.Post("", authMiddleware, UploadImage)
}

func UploadImage(c *fiber.Ctx) error {
	// check for credentials
	_, err := jwt.GetLoggedInUser(c)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(model.ErrorResponse{
			Message: err.Error(),
			Code:    "forbidden",
		})
	}

	fileReader, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "uploaded file was invalid",
			Code:    "invalid_file",
		})
	}

	// check file size & extension
	fileSize := fileReader.Size
	if fileSize < 10*1024 || fileSize > 2*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid file size",
			Code:    "invalid_file_size",
		})
	}

	// check extension
	fp := strings.Split(fileReader.Filename, ".")
	if len(fp) < 2 || (fp[len(fp)-1] != "jpg" && fp[len(fp)-1] != "jpeg") {
		return c.Status(fiber.StatusBadRequest).JSON(model.ErrorResponse{
			Message: "invalid file extension",
			Code:    "invalid_file_extension",
		})
	}

	imgUrl, err := S3ProviderImpl.UploadImage(c.Context(), fileReader)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(model.ErrorResponse{
			Message: "something wrong with the server. Please contact admin",
			Code:    "internal_server_error",
		})
	}

	return c.Status(fiber.StatusOK).JSON(ImageUploadResponse{
		ImageURL: imgUrl,
	})
}
