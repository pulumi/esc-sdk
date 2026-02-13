// Copyright 2024, Pulumi Corporation.  All rights reserved.

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// operationIDMap maps original operationIds from the centralized spec to the
// existing SDK method names for backward compatibility.
var operationIDMap = map[string]string{
	"ListOrgEnvironments_esc":                         "ListEnvironments",
	"CreateEnvironment_esc_environments":              "CreateEnvironment",
	"ReadEnvironment_esc_environments":                "GetEnvironment",
	"UpdateEnvironment_esc_environments":              "UpdateEnvironmentYaml",
	"HeadEnvironment_esc_environments":                "GetEnvironmentETag",
	"DeleteEnvironment_esc_environments":              "DeleteEnvironment",
	"CheckYAML_esc":                                   "CheckEnvironmentYaml",
	"DecryptEnvironment_esc_environments":             "DecryptEnvironment",
	"OpenEnvironment_esc_environments":                "OpenEnvironment",
	"ReadOpenEnvironment_esc_environments":            "ReadOpenEnvironment",
	"OpenEnvironment_esc_environments_versions":       "OpenEnvironmentAtVersion",
	"ReadEnvironment_esc_environments_versions":       "GetEnvironmentAtVersion",
	"ListEnvironmentTags_esc_environments":            "ListEnvironmentTags",
	"CreateEnvironmentTag_esc_environments":           "CreateEnvironmentTag",
	"GetEnvironmentTag_esc_environments":              "GetEnvironmentTag",
	"UpdateEnvironmentTag_esc_environments":           "UpdateEnvironmentTag",
	"DeleteEnvironmentTag_esc_environments":           "DeleteEnvironmentTag",
	"ListEnvironmentRevisions_esc_environments":       "ListEnvironmentRevisions",
	"ListRevisionTags_esc_environments_versions":      "ListEnvironmentRevisionTags",
	"CreateRevisionTag_esc_environments_versions_tags": "CreateEnvironmentRevisionTag",
	"ReadRevisionTag_esc_environments":                "GetEnvironmentRevisionTag",
	"UpdateRevisionTag_esc_environments":              "UpdateEnvironmentRevisionTag",
	"DeleteRevisionTag_esc_environments":              "DeleteEnvironmentRevisionTag",
	"CloneEnvironment":                                "CloneEnvironment",
}

// suffixesToStrip are tried in order (longest first) when an operationId is
// not in the explicit map.
var suffixesToStrip = []string{
	"_esc_environments_versions_tags",
	"_esc_environments_versions",
	"_esc_environments",
	"_esc",
}

func main() {
	input := flag.String("input", "https://api.pulumi.com/api/openapi/pulumi-spec.json", "URL or local path to input OpenAPI spec (JSON)")
	output := flag.String("output", "sdk/swagger.yaml", "Output path for filtered YAML spec")
	flag.Parse()

	spec, err := fetchSpec(*input)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching spec: %v\n", err)
		os.Exit(1)
	}

	filtered, err := filterSpec(spec)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error filtering spec: %v\n", err)
		os.Exit(1)
	}

	if err := writeYAML(*output, filtered); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Wrote filtered spec to %s\n", *output)
}

// fetchSpec loads the spec from a URL or local file path.
func fetchSpec(input string) (map[string]any, error) {
	var data []byte
	var err error

	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		resp, err := http.Get(input)
		if err != nil {
			return nil, fmt.Errorf("HTTP GET failed: %w", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("HTTP %d from %s", resp.StatusCode, input)
		}
		data, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("reading response body: %w", err)
		}
	} else {
		data, err = os.ReadFile(input)
		if err != nil {
			return nil, fmt.Errorf("reading file: %w", err)
		}
	}

	var spec map[string]any
	if err := json.Unmarshal(data, &spec); err != nil {
		return nil, fmt.Errorf("parsing JSON: %w", err)
	}
	return spec, nil
}

// filterSpec applies all transformation steps to produce the filtered spec.
func filterSpec(spec map[string]any) (map[string]any, error) {
	paths := getMap(spec, "paths")
	if paths == nil {
		return nil, fmt.Errorf("no paths found in spec")
	}

	// Step 1: Filter to /api/esc/ paths, excluding /api/preview/
	filtered := filterPaths(paths)

	// Step 2: Strip /api/esc prefix
	stripped := stripPrefix(filtered)

	// Step 3-6: Remap operationIds, force tags, apply security, fix responses.
	// Two passes: first assign explicitly mapped IDs, then handle suffix stripping
	// with deduplication.
	usedIDs := make(map[string]bool)

	// First pass: apply explicit map entries
	for pathKey, pathItem := range stripped {
		pi, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
			op, ok := pi[method].(map[string]any)
			if !ok {
				continue
			}
			if origID, ok := op["operationId"].(string); ok {
				if mapped, exists := operationIDMap[origID]; exists {
					op["operationId"] = mapped
					usedIDs[mapped] = true
				}
			}
			// Force tags
			op["tags"] = []any{"esc"}
			// Apply security
			op["security"] = []any{map[string]any{"Authorization": []any{}}}

			pi[method] = op
		}
		stripped[pathKey] = pi
	}

	// Second pass: apply suffix stripping for operations not in the explicit map,
	// avoiding collisions with already-assigned IDs.
	for pathKey, pathItem := range stripped {
		pi, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
			op, ok := pi[method].(map[string]any)
			if !ok {
				continue
			}
			origID, ok := op["operationId"].(string)
			if !ok {
				continue
			}
			// Skip if already remapped via explicit map
			if usedIDs[origID] {
				continue
			}
			// Try suffix stripping
			newID := origID
			for _, suffix := range suffixesToStrip {
				if strings.HasSuffix(origID, suffix) {
					candidate := strings.TrimSuffix(origID, suffix)
					if !usedIDs[candidate] {
						newID = candidate
					}
					break
				}
			}
			op["operationId"] = newID
			usedIDs[newID] = true
			pi[method] = op
		}
		stripped[pathKey] = pi
	}

	// Step 6.5: Fix missing request bodies in centralized spec
	fixMissingRequestBodies(stripped)

	// Step 7: Override ReadOpenEnvironment response to return EscEnvironment
	overrideReadOpenEnvironmentResponse(stripped)

	// Step 8: Synthesize ReadOpenEnvironmentProperty
	synthesizeReadOpenEnvironmentProperty(stripped)

	// Step 9: Collect schemas transitively from $ref pointers
	components := getMap(spec, "components")
	allSchemas := getMap(components, "schemas")
	allParameters := getMap(components, "parameters")
	allResponses := getMap(components, "responses")

	referencedSchemas := collectReferencedSchemas(stripped, allSchemas)

	// Step 9.5: Fix codegen conflicts and schema overrides
	fixCodegenConflicts(referencedSchemas)
	fixSchemaTypes(referencedSchemas)

	// Step 10: Inject local schemas
	injectLocalSchemas(referencedSchemas)

	// Step 11: Collect referenced parameters and responses
	referencedParams := collectReferencedComponents(stripped, allParameters, "parameters")
	referencedResps := collectReferencedComponents(stripped, allResponses, "responses")

	// Also collect schemas referenced by the collected responses and parameters
	collectTransitiveSchemasFromComponents(referencedResps, allSchemas, referencedSchemas)
	collectTransitiveSchemasFromComponents(referencedParams, allSchemas, referencedSchemas)

	// Step 12: Build output
	output := buildOutput(stripped, referencedSchemas, referencedParams, referencedResps)

	return output, nil
}

// filterPaths returns only paths with /api/esc/ prefix, excluding /api/preview/.
func filterPaths(paths map[string]any) map[string]any {
	result := make(map[string]any)
	for path, item := range paths {
		if strings.HasPrefix(path, "/api/esc/") && !strings.HasPrefix(path, "/api/preview/") {
			result[path] = item
		}
	}
	return result
}

// stripPrefix removes /api/esc from path keys.
func stripPrefix(paths map[string]any) map[string]any {
	result := make(map[string]any)
	for path, item := range paths {
		newPath := strings.TrimPrefix(path, "/api/esc")
		result[newPath] = item
	}
	return result
}

// fixMissingRequestBodies adds request bodies that are missing from the centralized spec
// but required for backward compatibility.
func fixMissingRequestBodies(paths map[string]any) {
	yamlBody := map[string]any{
		"description": "Environment Yaml content",
		"required":    true,
		"content": map[string]any{
			"application/x-yaml": map[string]any{
				"schema": map[string]any{
					"type": "string",
				},
			},
		},
	}

	for _, pathItem := range paths {
		pi, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range []string{"get", "post", "put", "patch", "delete"} {
			op, ok := pi[method].(map[string]any)
			if !ok {
				continue
			}
			opID, _ := op["operationId"].(string)
			switch opID {
			case "CheckEnvironmentYaml":
				if _, hasBody := op["requestBody"]; !hasBody {
					op["requestBody"] = yamlBody
				}
			}
		}
	}
}

// overrideReadOpenEnvironmentResponse changes the ReadOpenEnvironment 200 response
// to return EscEnvironment instead of a generic object.
func overrideReadOpenEnvironmentResponse(paths map[string]any) {
	for _, pathItem := range paths {
		pi, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head"} {
			op, ok := pi[method].(map[string]any)
			if !ok {
				continue
			}
			opID, _ := op["operationId"].(string)
			if opID != "ReadOpenEnvironment" {
				continue
			}
			responses := getMap(op, "responses")
			if responses == nil {
				continue
			}
			responses["200"] = map[string]any{
				"description": "Success",
				"content": map[string]any{
					"application/json": map[string]any{
						"schema": map[string]any{
							"$ref": "#/components/schemas/EscEnvironment",
						},
					},
				},
			}
			op["responses"] = responses
		}
	}
}

// synthesizeReadOpenEnvironmentProperty creates the double-slash path for property reads.
func synthesizeReadOpenEnvironmentProperty(paths map[string]any) {
	// Find the ReadOpenEnvironment path to copy parameters from
	var readOpenPath string
	var readOpenPathItem map[string]any
	for pathKey, pathItem := range paths {
		pi, ok := pathItem.(map[string]any)
		if !ok {
			continue
		}
		getOp, ok := pi["get"].(map[string]any)
		if !ok {
			continue
		}
		if opID, _ := getOp["operationId"].(string); opID == "ReadOpenEnvironment" {
			readOpenPath = pathKey
			readOpenPathItem = pi
			break
		}
	}

	if readOpenPathItem == nil {
		fmt.Fprintf(os.Stderr, "Warning: ReadOpenEnvironment path not found, cannot synthesize ReadOpenEnvironmentProperty\n")
		return
	}

	// Build new path with double-slash
	// Original: /environments/{orgName}/{projectName}/{envName}/open/{openSessionID}
	// New:      /environments/{orgName}/{projectName}/{envName}/open//{openSessionID}
	newPath := strings.Replace(readOpenPath, "/open/{open", "/open//{open", 1)

	// Collect path parameters from both path-level and operation-level.
	// The centralized spec may define path params at the operation level.
	var pathParams []any
	if params, ok := readOpenPathItem["parameters"].([]any); ok {
		for _, p := range params {
			if pm, ok := p.(map[string]any); ok {
				if pm["in"] == "path" {
					pathParams = append(pathParams, p)
				}
			}
		}
	}
	getOp := readOpenPathItem["get"].(map[string]any)
	if opParams, ok := getOp["parameters"].([]any); ok {
		for _, p := range opParams {
			if pm, ok := p.(map[string]any); ok {
				if pm["in"] == "path" {
					pathParams = append(pathParams, p)
				}
			}
		}
	}

	// Add property query parameter
	pathParams = append(pathParams, map[string]any{
		"name":        "property",
		"in":          "query",
		"required":    true,
		"description": "Path to a specific property using Pulumi path syntax https://www.pulumi.com/docs/concepts/config/#structured-configuration",
		"schema": map[string]any{
			"type": "string",
		},
	})

	newPathItem := map[string]any{
		"parameters": pathParams,
		"get": map[string]any{
			"tags":        []any{"esc"},
			"operationId": "ReadOpenEnvironmentProperty",
			"summary":     "Read an open environment property",
			"description": "Reads and decrypts a specific property including retrieving dynamic secrets from providers.",
			"security":    []any{map[string]any{"Authorization": []any{}}},
			"responses": map[string]any{
				"200": map[string]any{
					"description": "Success",
					"content": map[string]any{
						"application/json": map[string]any{
							"schema": map[string]any{
								"$ref": "#/components/schemas/EscValue",
							},
						},
					},
				},
				"401": map[string]any{
					"$ref": "#/components/responses/Unauthorized",
				},
				"404": map[string]any{
					"$ref": "#/components/responses/NotFound",
				},
				"500": map[string]any{
					"$ref": "#/components/responses/InternalServerError",
				},
				"default": map[string]any{
					"$ref": "#/components/responses/InternalServerError",
				},
			},
		},
	}

	paths[newPath] = newPathItem
}

// collectReferencedSchemas walks all $ref pointers in the filtered paths and
// collects the transitive closure of referenced schemas.
func collectReferencedSchemas(paths map[string]any, allSchemas map[string]any) map[string]any {
	refs := collectRefs(paths)

	needed := make(map[string]bool)
	for _, ref := range refs {
		name := schemaNameFromRef(ref)
		if name != "" {
			needed[name] = true
		}
	}

	// Transitively collect
	result := make(map[string]any)
	visited := make(map[string]bool)
	var resolve func(name string)
	resolve = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		schema, ok := allSchemas[name]
		if !ok {
			return
		}
		result[name] = schema
		// Find nested refs
		nestedRefs := collectRefs(schema)
		for _, ref := range nestedRefs {
			n := schemaNameFromRef(ref)
			if n != "" {
				resolve(n)
			}
		}
	}

	for name := range needed {
		resolve(name)
	}

	return result
}

// collectTransitiveSchemasFromComponents walks collected components (responses/parameters)
// and adds any schemas they reference to the referencedSchemas map.
func collectTransitiveSchemasFromComponents(components map[string]any, allSchemas map[string]any, referencedSchemas map[string]any) {
	refs := collectRefs(components)
	visited := make(map[string]bool)
	for _, existing := range sortedKeys(referencedSchemas) {
		visited[existing] = true
	}

	var resolve func(name string)
	resolve = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		schema, ok := allSchemas[name]
		if !ok {
			return
		}
		referencedSchemas[name] = schema
		nestedRefs := collectRefs(schema)
		for _, ref := range nestedRefs {
			n := schemaNameFromRef(ref)
			if n != "" {
				resolve(n)
			}
		}
	}

	for _, ref := range refs {
		name := schemaNameFromRef(ref)
		if name != "" {
			resolve(name)
		}
	}
}

// collectRefs recursively finds all $ref string values in the data.
func collectRefs(data any) []string {
	var refs []string
	var walk func(v any)
	walk = func(v any) {
		switch val := v.(type) {
		case map[string]any:
			if ref, ok := val["$ref"].(string); ok {
				refs = append(refs, ref)
			}
			for _, child := range val {
				walk(child)
			}
		case []any:
			for _, child := range val {
				walk(child)
			}
		}
	}
	walk(data)
	return refs
}

// schemaNameFromRef extracts the schema name from a $ref like "#/components/schemas/Foo".
func schemaNameFromRef(ref string) string {
	const prefix = "#/components/schemas/"
	if strings.HasPrefix(ref, prefix) {
		return strings.TrimPrefix(ref, prefix)
	}
	return ""
}

// collectReferencedComponents collects parameters or responses referenced by $ref
// from the filtered paths.
func collectReferencedComponents(paths map[string]any, allComponents map[string]any, kind string) map[string]any {
	if allComponents == nil {
		return make(map[string]any)
	}
	prefix := "#/components/" + kind + "/"
	refs := collectRefs(paths)

	result := make(map[string]any)
	visited := make(map[string]bool)

	var resolve func(name string)
	resolve = func(name string) {
		if visited[name] {
			return
		}
		visited[name] = true
		comp, ok := allComponents[name]
		if !ok {
			return
		}
		result[name] = comp
		// Also collect any nested refs to the same component kind
		nestedRefs := collectRefs(comp)
		for _, ref := range nestedRefs {
			if strings.HasPrefix(ref, prefix) {
				n := strings.TrimPrefix(ref, prefix)
				resolve(n)
			}
		}
	}

	for _, ref := range refs {
		if strings.HasPrefix(ref, prefix) {
			name := strings.TrimPrefix(ref, prefix)
			resolve(name)
		}
	}

	return result
}

// fixCodegenConflicts renames schema properties that conflict with generated method names
// in Go codegen (e.g., a field "hasSecret" collides with the generated "HasSecret()" method
// for the optional "secret" field).
func fixCodegenConflicts(schemas map[string]any) {
	if webhook, ok := schemas["WebhookResponse"].(map[string]any); ok {
		renameProperty(webhook, "hasSecret", "hasSecretValue")
	}
}

// renameProperty renames a property in a schema, handling both direct properties
// and properties inside allOf compositions.
func renameProperty(schema map[string]any, oldName, newName string) {
	doRename := func(obj map[string]any) {
		if props, ok := obj["properties"].(map[string]any); ok {
			if val, exists := props[oldName]; exists {
				props[newName] = val
				delete(props, oldName)
			}
		}
		if req, ok := obj["required"].([]any); ok {
			for i, r := range req {
				if r == oldName {
					req[i] = newName
				}
			}
		}
	}

	// Try direct properties
	doRename(schema)

	// Try allOf
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, item := range allOf {
			if m, ok := item.(map[string]any); ok {
				doRename(m)
			}
		}
	}
}

// fixSchemaTypes corrects schema property types that differ between the centralized spec
// and what the SDK extensions expect. For example, the centralized spec may define a
// free-form value field as "type: object" when it should be truly free-form (any).
func fixSchemaTypes(schemas map[string]any) {
	// EscValue.value must be free-form (any type), not just object
	setPropertyFreeForm(schemas, "EscValue", "value")

	// EscSchemaSchema: const, default, enum items, and examples items can be any
	// JSON value (string, number, bool, etc.), not just object.
	setPropertyFreeForm(schemas, "EscSchemaSchema", "const")
	setPropertyFreeForm(schemas, "EscSchemaSchema", "default")
	setPropertyArrayItemsFreeForm(schemas, "EscSchemaSchema", "enum")
	setPropertyArrayItemsFreeForm(schemas, "EscSchemaSchema", "examples")

	// EscExpr.literal can be any JSON value (string, number, bool, null), not object.
	setPropertyFreeForm(schemas, "EscExpr", "literal")

	// EscSchemaSchema is a JSON Schema representation. In JSON Schema, any schema
	// position can be a boolean (true = accept all, false = reject all) or an object.
	// The API returns booleans for many schema-typed fields. Make ALL references to
	// EscSchemaSchema free-form across all schemas.
	makeAllSchemaRefsFreeForm(schemas, "#/components/schemas/EscSchemaSchema")

	// EnvironmentFunctionSummary.rotationPaths can be null from the API, so make it nullable.
	setPropertyNullable(schemas, "EnvironmentFunctionSummary", "rotationPaths")
}

// setPropertyNullable adds "nullable: true" to a schema property so that
// the generated code accepts null values. Works with direct properties and allOf.
func setPropertyNullable(schemas map[string]any, schemaName, propName string) {
	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		return
	}
	doFix := func(obj map[string]any) {
		if props, ok := obj["properties"].(map[string]any); ok {
			if prop, ok := props[propName].(map[string]any); ok {
				prop["nullable"] = true
			}
		}
	}
	doFix(schema)
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, item := range allOf {
			if m, ok := item.(map[string]any); ok {
				doFix(m)
			}
		}
	}
}

// setPropertyFreeForm removes the "type" constraint from a schema property,
// making it a free-form field (any type). Works with direct properties and allOf.
func setPropertyFreeForm(schemas map[string]any, schemaName, propName string) {
	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		return
	}
	doFix := func(obj map[string]any) {
		if props, ok := obj["properties"].(map[string]any); ok {
			if prop, ok := props[propName].(map[string]any); ok {
				delete(prop, "type")
			}
		}
	}
	doFix(schema)
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, item := range allOf {
			if m, ok := item.(map[string]any); ok {
				doFix(m)
			}
		}
	}
}

// setPropertyArrayItemsFreeForm removes the "type" constraint from items of an
// array property, making each item free-form (any type).
func setPropertyArrayItemsFreeForm(schemas map[string]any, schemaName, propName string) {
	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		return
	}
	doFix := func(obj map[string]any) {
		if props, ok := obj["properties"].(map[string]any); ok {
			if prop, ok := props[propName].(map[string]any); ok {
				if items, ok := prop["items"].(map[string]any); ok {
					delete(items, "type")
				}
			}
		}
	}
	doFix(schema)
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, item := range allOf {
			if m, ok := item.(map[string]any); ok {
				doFix(m)
			}
		}
	}
}

// removePropertyRef replaces a $ref property with a free-form field (no type/ref constraints),
// useful when a field can be either a typed object or a boolean/primitive (e.g. JSON Schema's items).
func removePropertyRef(schemas map[string]any, schemaName, propName string) {
	schema, ok := schemas[schemaName].(map[string]any)
	if !ok {
		return
	}
	doFix := func(obj map[string]any) {
		if props, ok := obj["properties"].(map[string]any); ok {
			if prop, ok := props[propName].(map[string]any); ok {
				delete(prop, "$ref")
				delete(prop, "type")
			}
		}
	}
	doFix(schema)
	if allOf, ok := schema["allOf"].([]any); ok {
		for _, item := range allOf {
			if m, ok := item.(map[string]any); ok {
				doFix(m)
			}
		}
	}
}

// makeAllSchemaRefsFreeForm finds every property across all schemas that uses the
// given $ref and removes the $ref, making the property free-form (interface{} in Go).
// This is needed for EscSchemaSchema references because JSON Schema positions can be
// booleans (true/false) which can't be deserialized into typed structs.
func makeAllSchemaRefsFreeForm(schemas map[string]any, ref string) {
	clearRef := func(props map[string]any) {
		for _, prop := range props {
			p, ok := prop.(map[string]any)
			if !ok {
				continue
			}
			// Direct $ref
			if p["$ref"] == ref {
				delete(p, "$ref")
			}
			// Array items with $ref
			if items, ok := p["items"].(map[string]any); ok {
				if items["$ref"] == ref {
					delete(items, "$ref")
				}
			}
			// additionalProperties with $ref
			if ap, ok := p["additionalProperties"].(map[string]any); ok {
				if ap["$ref"] == ref {
					delete(ap, "$ref")
				}
			}
		}
	}
	for _, schema := range schemas {
		s, ok := schema.(map[string]any)
		if !ok {
			continue
		}
		if props, ok := s["properties"].(map[string]any); ok {
			clearRef(props)
		}
		if allOf, ok := s["allOf"].([]any); ok {
			for _, item := range allOf {
				if m, ok := item.(map[string]any); ok {
					if props, ok := m["properties"].(map[string]any); ok {
						clearRef(props)
					}
				}
			}
		}
	}
}

// injectLocalSchemas adds schemas that don't exist in the new spec but are needed by extensions.
func injectLocalSchemas(schemas map[string]any) {
	if _, ok := schemas["EnvironmentDefinition"]; !ok {
		schemas["EnvironmentDefinition"] = map[string]any{
			"type": "object",
			"properties": map[string]any{
				"imports": map[string]any{
					"type": "array",
					"items": map[string]any{
						"type": "string",
					},
				},
				"values": map[string]any{
					"$ref": "#/components/schemas/EnvironmentDefinitionValues",
				},
			},
			"example": map[string]any{
				"application/x-yaml": map[string]any{
					"imports": []any{"base-env"},
					"values": map[string]any{
						"foo": "bar",
					},
					"pulumiConfig": map[string]any{
						"foo": "${foo}",
					},
					"environmentVariables": map[string]any{
						"MY_KEY": "my-value",
					},
				},
			},
		}
	}
	if _, ok := schemas["EnvironmentDefinitionValues"]; !ok {
		schemas["EnvironmentDefinitionValues"] = map[string]any{
			"type": "object",
			"properties": map[string]any{
				"pulumiConfig": map[string]any{
					"type":                 "object",
					"additionalProperties": true,
				},
				"environmentVariables": map[string]any{
					"type": "object",
					"additionalProperties": map[string]any{
						"type": "string",
					},
				},
				"files": map[string]any{
					"type": "object",
					"additionalProperties": map[string]any{
						"type": "string",
					},
				},
			},
			"additionalProperties": map[string]any{
				"type": "object",
			},
		}
	}
	if _, ok := schemas["Error"]; !ok {
		schemas["Error"] = map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{"type": "string"},
				"code":    map[string]any{"type": "integer"},
			},
			"required": []any{"message", "code"},
		}
	}
	if _, ok := schemas["Reference"]; !ok {
		schemas["Reference"] = map[string]any{
			"type":     "object",
			"required": []any{"$ref"},
			"properties": map[string]any{
				"$ref": map[string]any{
					"type":   "string",
					"format": "uri-reference",
				},
			},
		}
	}
}

// buildOutput assembles the final OpenAPI spec.
func buildOutput(paths, schemas, parameters, responses map[string]any) map[string]any {
	components := map[string]any{
		"securitySchemes": map[string]any{
			"Authorization": map[string]any{
				"type": "apiKey",
				"name": "Authorization",
				"in":   "header",
			},
		},
		"schemas": schemas,
	}
	if len(parameters) > 0 {
		components["parameters"] = parameters
	}
	if len(responses) > 0 {
		components["responses"] = responses
	}

	return map[string]any{
		"openapi": "3.0.3",
		"info": map[string]any{
			"title":       "ESC (Environments, Secrets, Config) API",
			"description": "Pulumi ESC allows you to compose and manage hierarchical collections of configuration and secrets and consume them in various ways.",
			"version":     "0.1.0",
			"license": map[string]any{
				"name": "Apache 2.0",
				"url":  "https://www.apache.org/licenses/LICENSE-2.0.html",
			},
		},
		"servers": []any{
			map[string]any{
				"url":         "https://api.pulumi.com/api/esc",
				"description": "Pulumi Cloud Production Preview API",
			},
		},
		"components": components,
		"security": []any{
			map[string]any{"Authorization": []any{}},
		},
		"paths": paths,
	}
}

// writeYAML writes the spec as YAML to the output path.
func writeYAML(outputPath string, spec map[string]any) error {
	if dir := filepath.Dir(outputPath); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating output directory: %w", err)
		}
	}

	// Build an ordered yaml.Node for stable output
	node := toOrderedYAMLNode(spec)
	docNode := &yaml.Node{
		Kind: yaml.DocumentNode,
		Content: []*yaml.Node{
			node,
		},
		HeadComment: "Generated by tools/filter-spec from the centralized Pulumi OpenAPI spec.\n# Do not edit manually.",
	}

	f, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("creating file: %w", err)
	}
	defer f.Close()

	// Write the YAML document separator
	if _, err := f.WriteString("---\n"); err != nil {
		return err
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(2)
	if err := enc.Encode(docNode); err != nil {
		return fmt.Errorf("encoding YAML: %w", err)
	}
	return enc.Close()
}

// toOrderedYAMLNode converts a Go value to a yaml.Node with sorted map keys.
func toOrderedYAMLNode(v any) *yaml.Node {
	switch val := v.(type) {
	case map[string]any:
		node := &yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}
		keys := sortedKeys(val)
		for _, k := range keys {
			keyNode := &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!str",
				Value: k,
			}
			valNode := toOrderedYAMLNode(val[k])
			node.Content = append(node.Content, keyNode, valNode)
		}
		return node
	case []any:
		node := &yaml.Node{
			Kind: yaml.SequenceNode,
			Tag:  "!!seq",
		}
		for _, item := range val {
			node.Content = append(node.Content, toOrderedYAMLNode(item))
		}
		return node
	case string:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: val,
		}
	case float64:
		// JSON numbers are float64
		if val == float64(int64(val)) {
			return &yaml.Node{
				Kind:  yaml.ScalarNode,
				Tag:   "!!int",
				Value: fmt.Sprintf("%d", int64(val)),
			}
		}
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!float",
			Value: fmt.Sprintf("%g", val),
		}
	case bool:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: fmt.Sprintf("%t", val),
		}
	case nil:
		return &yaml.Node{
			Kind: yaml.ScalarNode,
			Tag:  "!!null",
		}
	case json.Number:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Value: val.String(),
		}
	default:
		return &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: fmt.Sprintf("%v", val),
		}
	}
}

// getMap safely retrieves a map[string]any from a parent map.
func getMap(m map[string]any, key string) map[string]any {
	if m == nil {
		return nil
	}
	v, ok := m[key]
	if !ok {
		return nil
	}
	result, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	return result
}

// sortedKeys returns the keys of a map sorted alphabetically.
func sortedKeys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
