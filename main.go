package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"gocv.io/x/gocv"
)

// convertPDFToPNG uses pdftoppm CLI to convert the first page of a PDF to a PNG file.
// Output is saved as {outputPrefix}.png in the same directory as the PDF.
func convertPDFToPNG(pdfPath, outputPrefix string) (string, error) {
	// Example: pdftoppm -png -singlefile input.pdf output
	cmd := exec.Command("pdftoppm", "-png", "-singlefile", pdfPath, outputPrefix)
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("pdftoppm error: %v", err)
	}

	// The resulting file will be something like outputPrefix.png
	dir := filepath.Dir(pdfPath)
	outputFile := filepath.Join(dir, outputPrefix+".png")
	return outputFile, nil
}

// extractSignature loads an image via gocv, thresholds it, finds the largest contour,
// crops it, and returns a Mat containing just the signature region.
func extractSignature(imgPath string) (gocv.Mat, error) {
	// Read image in color
	img := gocv.IMRead(imgPath, gocv.IMReadColor)
	if img.Empty() {
		return gocv.NewMat(), fmt.Errorf("unable to read image: %s", imgPath)
	}
	defer img.Close()

	// Convert to grayscale
	gray := gocv.NewMat()
	gocv.CvtColor(img, &gray, gocv.ColorBGRToGray)
	defer gray.Close()

	// Threshold: convert signature (dark) to white, background (light) to black
	//   Adjust threshold (200) as needed for your scans
	bin := gocv.NewMat()
	// We use ThresholdBinaryInv so that dark ink becomes white (255)
	// and light background becomes black (0).
	gocv.Threshold(gray, &bin, 200, 255, gocv.ThresholdBinaryInv)
	defer bin.Close()

	// Find external contours
	contours := gocv.FindContours(bin, gocv.RetrievalExternal, gocv.ChainApproxSimple)
	defer contours.Close()

	// If there are no contours, we can't find a signature
	if contours.Size() == 0 {
		return gocv.NewMat(), fmt.Errorf("no contours found - cannot find signature")
	}

	// Find largest contour by bounding-rectangle area
	var maxArea float64
	var maxRect image.Rectangle

	// Iterate over the contours in the PointsVector
	for i := 0; i < contours.Size(); i++ {
		c := contours.At(i)          // c is of type gocv.Points
		rect := gocv.BoundingRect(c) // bounding box of this contour
		area := float64(rect.Dx() * rect.Dy())

		if area > maxArea {
			maxArea = area
			maxRect = rect
		}
	}

	// Crop the largest contour area from the original color image (img)
	signature := img.Region(maxRect)

	// Return a copy so we can safely Close() signature
	signatureCopy := signature.Clone()
	signature.Close()

	return signatureCopy, nil
}

// removeWhiteBackground converts near-white pixels to transparent (alpha=0)
// and keeps signature pixels opaque.
func removeWhiteBackground(input gocv.Mat) (image.Image, error) {
	// input is a BGR image (3 channels).
	if input.Channels() != 3 {
		return nil, fmt.Errorf("expected 3-channel BGR image")
	}

	rows := input.Rows()
	cols := input.Cols()

	// Create a new RGBA image in Go
	output := image.NewRGBA(image.Rect(0, 0, cols, rows))

	// Read each pixel, if it's near white => make alpha=0, else alpha=255
	for y := 0; y < rows; y++ {
		for x := 0; x < cols; x++ {
			bVec := input.GetVecbAt(y, x)
			// bVec[0] = Blue, bVec[1] = Green, bVec[2] = Red
			b := bVec[0]
			g := bVec[1]
			r := bVec[2]

			// Simple "near-white" threshold
			if r > 200 && g > 200 && b > 200 {
				// transparent
				output.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 0})
			} else {
				// opaque
				output.Set(x, y, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	return output, nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <path_to_pdf>")
		return
	}

	pdfPath := os.Args[1]
	fmt.Printf("Converting PDF: %s\n", pdfPath)

	// Step 1: Convert first page of PDF to PNG
	outputPrefix := "pdf_page"
	pngPath, err := convertPDFToPNG(pdfPath, outputPrefix)
	if err != nil {
		log.Fatalf("Failed to convert PDF to PNG: %v", err)
	}

	fmt.Printf("PNG generated: %s\n", pngPath)

	// Step 2: Extract signature region
	signatureMat, err := extractSignature(pngPath)
	if err != nil {
		log.Fatalf("Failed to extract signature: %v", err)
	}
	defer signatureMat.Close()

	// Step 3: Remove white background (convert near-white to transparent)
	signatureImage, err := removeWhiteBackground(signatureMat)
	if err != nil {
		log.Fatalf("Failed to remove background: %v", err)
	}

	// Step 4: Save final PNG
	outFile, err := os.Create("signature_result.png")
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, signatureImage); err != nil {
		log.Fatalf("Failed to encode PNG: %v", err)
	}

	fmt.Println("Signature with transparent background saved to signature_result.png")
}
