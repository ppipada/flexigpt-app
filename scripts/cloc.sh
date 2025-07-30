#!/bin/bash

cloc --vcs=git --ignored=ignored.txt --exclude_ext=yaml,sum,mod .

# ===
# Output on 16 Apr 2025

#      247 text files.
#      233 unique files.
# Wrote ignored.txt
#       14 files ignored.

# github.com/AlDanial/cloc v 2.04  T=0.22 s (1042.1 files/s, 156292.5 lines/s)
# -------------------------------------------------------------------------------
# Language                     files          blank        comment           code
# -------------------------------------------------------------------------------
# Go                             121           2008           1539          15589
# TypeScript                      71           1176           1111           8368
# JSON                            12              8              0           2477
# Markdown                         6            239              0            738
# YAML                             7            158            146            550
# JavaScript                       6             77             48            289
# Bourne Shell                     6             49             39            186
# XML                              2              9              1             75
# CSS                              1              7              7             36
# INI                              1              2              0             12
# -------------------------------------------------------------------------------
# SUM:                           233           3733           2891          28320
# -------------------------------------------------------------------------------

# ignored.txt

# ./.gitattributes:  listed in $Not_Code_Extension{gitattributes}
# ./.gitignore:  listed in $Not_Code_Extension{gitignore}
# ./.gitmodules:  listed in $Not_Code_Extension{gitmodules}
# ./.npmrc:  language unknown (#3)
# ./.prettierignore:  language unknown (#3)
# ./LICENSE:  language unknown (#3)
# ./frontend/.env.dev:  language unknown (#3)
# ./frontend/.env.prod:  language unknown (#3)
# ./frontend/package.json.md5:  language unknown (#3)
# ./go.mod:  listed in $Not_Code_Extension{mod}
# ./go.sum:  listed in $Not_Code_Extension{sum}
# ./packaging/flatpak/io.github.flexigpt.client.desktop:  listed in $Not_Code_Extension{desktop}
# ./pnpm-lock.yaml:  listed in $Not_Code_Extension{yaml}
# ./pnpm-workspace.yaml:  listed in $Not_Code_Extension{yaml}
# ===

# ===
# Output on 14 Nov 2024

#      226 text files.
#      203 unique files.
# Wrote ignored.txt
#       23 files ignored.

# github.com/AlDanial/cloc v 1.98  T=0.18 s (1111.8 files/s, 113194.2 lines/s)
# -------------------------------------------------------------------------------
# Language                     files          blank        comment           code
# -------------------------------------------------------------------------------
# Go                              61           1059            505           7939
# TypeScript                      88           1021            377           7331
# JSON                            26              0              0            628
# Markdown                         6            196              0            552
# YAML                             8             48             11            438
# JavaScript                       7             70             43            218
# Bourne Shell                     3             27             24             84
# XML                              2              9              1             75
# INI                              1              1              0              7
# CSS                              1              0              0              3
# -------------------------------------------------------------------------------
# SUM:                           203           2431            961          17275
# -------------------------------------------------------------------------------

# ignored.txt

# .gitattributes:  language unknown (#3)
# .gitignore:  listed in $Not_Code_Extension{gitignore}
# .gitmodules:  language unknown (#3)
# .npmrc:  language unknown (#3)
# .prettierignore:  language unknown (#3)
# .prettierrc:  language unknown (#3)
# agentts/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# frontend/app/models/aiprovider/chat_types.ts:  duplicate of pkgts/aiprovider/spec/chat_types.ts
# frontend/app/models/conversation.ts:  duplicate of pkgts/conversationstore/conversation_types.ts
# frontend/package.json.md5:  language unknown (#3)
# go.mod:  listed in $Not_Code_Extension{mod}
# go.sum:  listed in $Not_Code_Extension{sum}
# packaging/flatpak/io.github.flexigpt.client.desktop:  listed in $Not_Code_Extension{desktop}
# pkgts/aiprovider/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# pkgts/aiprovider/tsconfig.json:  duplicate of pkgts/winstonlogger/tsconfig.json
# pkgts/conversationstore/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# pkgts/logger/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# pkgts/logger/index.ts:  duplicate of pkgts/winstonlogger/index.ts
# pkgts/securejsondb/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# pkgts/securejsondb/tsconfig.json:  duplicate of pkgts/winstonlogger/tsconfig.json
# pkgts/settingstore/.eslintrc.json:  duplicate of pkgts/winstonlogger/.eslintrc.json
# pnpm-lock.yaml:  listed in $Not_Code_Extension{yaml}
# pnpm-workspace.yaml:  listed in $Not_Code_Extension{yaml}
# ===
