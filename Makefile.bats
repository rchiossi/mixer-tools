SUBMAKEFILES := $(shell find bat/ -name Makefile)
DIRS2RUNMAKECHECK := $(addprefix checkdir-,${SUBMAKEFILES})
DIRS2RUNMAKECLEAN := $(addprefix clean-,${SUBMAKEFILES})

GROUP1 := $(addprefix checkdir-,$(shell find bat/tests1 -name Makefile))
GROUP2 := $(addprefix checkdir-,$(shell find bat/tests2 -name Makefile))
GROUP3 := $(addprefix checkdir-,$(shell find bat/tests3 -name Makefile))
GROUP4 := $(addprefix checkdir-,$(shell find bat/tests4 -name Makefile))

batcheck: ${GROUP1} ${GROUP2} ${GROUP3} ${GROUP4}

batcheck-1: ${GROUP1}

batcheck-2: ${GROUP2}

batcheck-3: ${GROUP3}

batcheck-4: ${GROUP4}

${GROUP1}: checkdir-%:
	$(MAKE) -C $(dir $(subst checkdir-,,$@)) check

${GROUP2}: checkdir-%:
	$(MAKE) -C $(dir $(subst checkdir-,,$@)) check

${GROUP3}: checkdir-%:
	$(MAKE) -C $(dir $(subst checkdir-,,$@)) check

${GROUP4}: checkdir-%:
	$(MAKE) -C $(dir $(subst checkdir-,,$@)) check

batclean: $(DIRS2RUNMAKECLEAN) clean-dnf

${DIRS2RUNMAKECLEAN}: clean-%:
	$(MAKE) -C $(dir $(subst clean-,,$@)) clean

clean-dnf:
	rm -rf bat/dnf

.PHONY: batcheck
.PHONY: batcheck-1
.PHONY: batcheck-2
.PHONY: batcheck-3
.PHONY: batclean-4
.PHONY: ${DIRS2RUNMAKECHECK}
.PHONY: ${DIRS2RUNMAKECLEAN}
