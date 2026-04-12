package render

import (
	"bytes"
	"strings"

	nethtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

func RewriteAssetLinks(htmlFragment string, assetURLs map[string]string) (string, error) {
	root, err := parseHTMLFragment(htmlFragment)
	if err != nil {
		return "", err
	}
	rewriteAssetNodes(root, assetURLs)
	return renderHTMLFragment(root)
}

func rewriteAssetNodes(node *nethtml.Node, assetURLs map[string]string) {
	if node.Type == nethtml.ElementNode {
		for index := range node.Attr {
			if node.Attr[index].Key != "src" && node.Attr[index].Key != "href" {
				continue
			}
			if target, exists := assetURLs[node.Attr[index].Val]; exists {
				node.Attr[index].Val = target
			}
		}
	}
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		rewriteAssetNodes(child, assetURLs)
	}
}

func parseHTMLFragment(fragment string) (*nethtml.Node, error) {
	container := &nethtml.Node{Type: nethtml.ElementNode, Data: "div", DataAtom: atom.Div}
	nodes, err := nethtml.ParseFragment(strings.NewReader(fragment), container)
	if err != nil {
		return nil, err
	}
	root := &nethtml.Node{Type: nethtml.ElementNode, Data: "div", DataAtom: atom.Div}
	for _, node := range nodes {
		root.AppendChild(node)
	}
	return root, nil
}

func renderHTMLFragment(root *nethtml.Node) (string, error) {
	var buffer bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		if err := nethtml.Render(&buffer, child); err != nil {
			return "", err
		}
	}
	return buffer.String(), nil
}
