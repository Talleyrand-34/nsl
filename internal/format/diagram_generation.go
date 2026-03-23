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

	"oss.terrastruct.com/d2/d2graph"
	"oss.terrastruct.com/d2/d2layouts/d2elklayout"
	"oss.terrastruct.com/d2/d2lib"
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

// WriteDiagram writes the D2 script and generated SVG image to files.
func WriteDiagram(
	d2Script string,
	outputPath string,
	outputDiagramFile string,
	outputImageFile string,
) {
	outFileD2 := filepath.Join(outputPath, outputDiagramFile)
	if err := os.WriteFile(outFileD2, []byte(d2Script), 0600); err != nil {
		panic("Failed to write D2lang script file: " + err.Error())
	}

	svgBytes, err := GenerateDiagramSVG(d2Script)
	if err != nil {
		panic(err)
	}

	outFilesvg := filepath.Join(outputPath, outputImageFile)
	if err := os.WriteFile(outFilesvg, svgBytes, 0600); err != nil {
		panic("Failed to write SVG file: " + err.Error())
	}

	fmt.Println("Diagram successfully generated in", outFilesvg)
}
