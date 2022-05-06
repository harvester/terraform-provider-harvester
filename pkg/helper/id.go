package helper

import (
	"fmt"
	"strings"
)

const (
	IDSep             = "/"
	NamespacedNameSep = "/"
)

func IDParts(id string) (string, string, error) {
	parts := strings.Split(id, IDSep)
	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	default:
		err := fmt.Errorf("unexpected ID format (%q), expected %q or %q. ", id, "namespace/name", "name")
		return "", "", err
	}
}

func BuildID(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + IDSep + name
}

func BuildNamespacedName(namespace, name string) string {
	if namespace == "" {
		return name
	}
	return namespace + NamespacedNameSep + name
}

func NamespacedNameParts(namespacedName string) (string, string, error) {
	parts := strings.Split(namespacedName, NamespacedNameSep)
	switch len(parts) {
	case 1:
		return "", parts[0], nil
	case 2:
		return parts[0], parts[1], nil
	default:
		err := fmt.Errorf("unexpected namespacedName format (%q), expected %q or %q. ", namespacedName, "namespace/name", "name")
		return "", "", err
	}
}

func NamespacedNamePartsByDefault(namespacedName string, defaultNamespace string) (string, string, error) {
	namespace, name, err := NamespacedNameParts(namespacedName)
	if err != nil {
		return "", "", err
	}
	if namespace == "" {
		namespace = defaultNamespace
	}
	return namespace, name, nil
}

func RebuildNamespacedName(namespacedName string, defaultNamespace string) (string, error) {
	namespace, name, err := NamespacedNamePartsByDefault(namespacedName, defaultNamespace)
	if err != nil {
		return "", err
	}
	return BuildNamespacedName(namespace, name), nil
}
