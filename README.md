# Go PDF Signature Extraction POC

This is a small proof-of-concept (POC) application written in **Go** that demonstrates how to:

1. Convert the first page of a PDF to a PNG file using **Poppler**’s `pdftoppm` CLI tool.
2. Use **GoCV** (OpenCV bindings for Go) to find and crop out a handwritten signature.
3. Remove the white background to produce a transparent PNG of the signature.

---

## Prerequisites

1. **Go** (1.18+ recommended)
2. **OpenCV 4** (native libraries, headers, etc.)
3. **GoCV** (Go bindings for OpenCV)
4. **Poppler** (for `pdftoppm` utility)

### Installing Prerequisites

<details>
<summary><strong>macOS (Homebrew)</strong></summary>

1. Install Go:

   ```bash
   brew install go
   ```

2. Install OpenCV:

   ```bash
   brew install opencv
   ```

3. Install Poppler:

   ```bash
   brew install poppler
   ```

4. Install GoCV (the Go module):

   ```bash
   go get -u gocv.io/x/gocv
   ```

5. Verify OpenCV installation:

   ```bash
   pkg-config --modversion opencv4
   ```

   **Expected output:** `4.5.3` (or a similar version).

</details>

<details>
<summary><strong>Ubuntu/Debian Linux</strong></summary>

1. Install Go (via apt or the official tarball from golang.org).

2. Install OpenCV dev packages & pkg-config:

   ```bash
   sudo apt-get update
   sudo apt-get install -y libopencv-dev pkg-config
   ```

3. Install Poppler:

   ```bash
   sudo apt-get install -y poppler-utils
   ```

4. Install GoCV:

   ```bash
   go get -u gocv.io/x/gocv
   ```

5. Verify OpenCV installation:

   ```bash
   pkg-config --modversion opencv4
   ```

   **Expected output:** `4.2.0`, `4.5.1`, etc.

</details>

<details>
<summary><strong>Windows (MSYS2 or vcpkg approach)</strong></summary>

Installing OpenCV + pkg-config on Windows can be more involved. The easiest method is to use MSYS2 or vcpkg.

**MSYS2 Example:**

1. Download and install MSYS2.

2. In the MSYS2 terminal, run:

   ```bash
   pacman -Syu
   pacman -S opencv pkg-config mingw-w64-x86_64-go
   ```

3. Install Poppler (or check if it's available in your MSYS2 environment).

4. Verify OpenCV installation:

   ```bash
   pkg-config --modversion opencv4
   ```

5. Install GoCV:

   ```bash
   go get -u gocv.io/x/gocv
   ```

</details>

---

## Project Structure

```
poc-pdf/
├── main.go
└── README.md
```

- `main.go`: Contains the code that converts a PDF to PNG, extracts the signature, and removes the background.
- `README.md`: This documentation file.

---

## Usage

1. Clone or copy this repository to your local machine.

2. Initialize a Go module (if not already done):

   ```bash
   go mod init poc-pdf
   go mod tidy
   ```

3. Run the proof-of-concept:

   ```bash
   go run main.go /path/to/your.pdf
   ```

   **Process:**

   - Converts the first page of `/path/to/your.pdf` to `pdf_page.png` using `pdftoppm`.
   - Uses GoCV to find and crop out the largest dark region (assumed to be the signature).
   - Removes white pixels (≥ 200 in R, G, B) by making them transparent.
   - Writes the result to `signature_result.png` in the current directory.

---

## How It Works

### PDF to Image Conversion

We use `pdftoppm` (part of Poppler) via `exec.Command("pdftoppm", ...)` to render the first page of a PDF to a PNG.

### Extract Signature (GoCV)

1. Load the PNG with `gocv.IMReadColor`.
2. Convert to grayscale.
3. Apply `ThresholdBinaryInv` (around 200). Dark pixels become white (255), background becomes black (0).
4. Find contours in the thresholded image.
5. Identify the largest bounding rectangle (assumed to be the signature).

### Remove White Background

1. Convert the cropped signature region (BGR) to a Go `image.RGBA`.
2. If the pixel is near-white (`r > 200 && g > 200 && b > 200`), set `alpha = 0` (transparent).
3. Otherwise, set `alpha = 255` (opaque).
4. Write the result to `signature_result.png`.

---

## Example

**Input:** `sample.pdf`  
**Command:**

```bash
go run main.go sample.pdf
```

**Output Files:**

- `pdf_page.png`: The extracted first page as a PNG.
- `signature_result.png`: The cropped signature with a transparent background.

---

## Troubleshooting

### Package 'opencv4' Not Found

- Ensure OpenCV 4 is installed.
- Verify with:

  ```bash
  pkg-config --modversion opencv4
  ```

### No Signature Found

- Adjust the threshold in `gocv.Threshold(...)`. Some PDFs might need `threshold=150` or `threshold=220`.
- Use morphological operations if the scan is noisy.

### Permissions / PATH Issues

- Ensure `pdftoppm` is on your system `PATH` or specify the full path in `exec.Command()`.
