# We have some template files that are basically Go
# source with some replacements. To make them more
# pleasant to edit we keep them as Go here and
# hack them into template format.

GOTARGETS := $(patsubst %/src.go,../template/db_%.tpl,$(wildcard */src.go))
TESTTARGETS := $(patsubst %/test.go,../template/test_%.tpl,$(wildcard */test.go))

.PHONY: all
all: $(GOTARGETS) $(TESTTARGETS)

../template/db_%.tpl: %/src.go
	sed -e 's/package .*/package {{ .Package }}/' -e 's/github\.com\/wttw\/schemadance/{{ .MigratePath }}/g' -e 's/schemadance/{{ .MigratePackage }}/g' $< >$@

../template/test_%.tpl: %/test.go
	sed -e 's/package .*/package {{ .Package }}/' -e 's/github\.com\/wttw\/schemadance/{{ .MigratePath }}/g' -e 's/schemadance/{{ .MigratePackage }}/g' $< >$@
