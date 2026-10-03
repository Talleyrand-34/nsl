// SPDX-License-Identifier: AGPL-3.0-or-later
/*
Copyright © 2025 Talleyrand-34 (t34@t34.dev)

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.
*/
package format

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
	"oss.terrastruct.com/d2/d2renderers/d2ascii"
	"oss.terrastruct.com/d2/d2renderers/d2ascii/charset"
	"oss.terrastruct.com/d2/d2renderers/d2svg"
	"oss.terrastruct.com/d2/d2themes/d2themescatalog"
	"oss.terrastruct.com/d2/lib/log"
	"oss.terrastruct.com/d2/lib/textmeasure"
	"oss.terrastruct.com/util-go/go2"
)

// GenerateDiagramSVG compiles the D2 script and returns the SVG image as bytes.
func GenerateDiagramSVG(d2Script string) ([]byte, error) {
	ruler, _ := textmeasure.NewRuler()
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(5)),
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}
	ctx := log.WithDefault(context.Background())

	opts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}
	graph, _, err := d2lib.Compile(ctx, d2Script, opts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to compile D2 script: %w", err)
	}

	outSvg, err := d2svg.Render(graph, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to render SVG: %w", err)
	}
	return outSvg, nil
}

// GenerateDiagramASCII compiles the D2 script and renders it as ASCII art using
// the in-process d2ascii renderer. With unicode=true it uses box-drawing
// characters; otherwise it falls back to plain ASCII (+ - |).
func GenerateDiagramASCII(d2Script string, unicode bool) ([]byte, error) {
	ruler, _ := textmeasure.NewRuler()
	layoutResolver := func(engine string) (d2graph.LayoutGraph, error) {
		return d2elklayout.DefaultLayout, nil
	}
	renderOpts := &d2svg.RenderOpts{
		Pad:     go2.Pointer(int64(5)),
		ThemeID: &d2themescatalog.GrapeSoda.ID,
	}
	ctx := log.WithDefault(context.Background())

	opts := &d2lib.CompileOptions{
		Ruler:          ruler,
		LayoutResolver: layoutResolver,
	}
	diagram, _, err := d2lib.Compile(ctx, d2Script, opts, renderOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to compile D2 script: %w", err)
	}

	charsetType := charset.Unicode
	if !unicode {
		charsetType = charset.ASCII
	}
	ascii, err := d2ascii.NewASCIIartist().Render(ctx, diagram, &d2ascii.RenderOpts{Charset: charsetType})
	if err != nil {
		return nil, fmt.Errorf("failed to render ASCII: %w", err)
	}
	return ascii, nil
}

// WriteDiagram writes the D2 script and renders the diagram. By default it
// renders an SVG image to outputImageFile. When ascii is set, it renders ASCII
// art instead: the art is printed to stdout and also written to a .txt file
// derived from outputImageFile (unicode toggles box-drawing vs plain ASCII).
func WriteDiagram(
	d2Script string,
	outputPath string,
	outputDiagramFile string,
	outputImageFile string,
	ascii bool,
	unicode bool,
) {
	outFileD2 := filepath.Join(outputPath, outputDiagramFile)
	if err := os.WriteFile(outFileD2, []byte(d2Script), 0600); err != nil {
		panic("Failed to write D2lang script file: " + err.Error())
	}

	if ascii {
		asciiBytes, err := GenerateDiagramASCII(d2Script, unicode)
		if err != nil {
			panic(err)
		}
		// Print to stdout (the diagram is the command's machine output).
		fmt.Print(string(asciiBytes))
		if len(asciiBytes) > 0 && asciiBytes[len(asciiBytes)-1] != '\n' {
			fmt.Println()
		}

		outFileTxt := filepath.Join(outputPath, asciiFileName(outputImageFile))
		if err := os.WriteFile(outFileTxt, asciiBytes, 0600); err != nil {
			panic("Failed to write ASCII file: " + err.Error())
		}
		fmt.Fprintln(os.Stderr, "Diagram successfully generated in", outFileTxt)
		return
	}

	svgBytes, err := GenerateDiagramSVG(d2Script)
	if err != nil {
		panic(err)
	}

	outFilesvg := filepath.Join(outputPath, outputImageFile)
	if err := os.WriteFile(outFilesvg, svgBytes, 0600); err != nil {
		panic("Failed to write SVG file: " + err.Error())
	}

	fmt.Fprintln(os.Stderr, "Diagram successfully generated in", outFilesvg)
}

// asciiFileName derives a .txt filename from the configured image filename so
// the ASCII output does not overwrite an SVG (e.g. out.svg -> out.txt).
func asciiFileName(imageFile string) string {
	ext := filepath.Ext(imageFile)
	if ext == "" {
		return imageFile + ".txt"
	}
	return strings.TrimSuffix(imageFile, ext) + ".txt"
}
