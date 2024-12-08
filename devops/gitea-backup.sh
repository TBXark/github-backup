#!/bin/bash

set -ex 
set -o pipefail 


GITEA_USER=git
GITEA_BIN=/home/git/gitea
GITEA_BACKUP_ZIP=/home/git/gitea-dump.zip
RCLONE_TARGET=gitea:Gitea/

[ -f $GITEA_BACKUP_ZIP ] && rm $GITEA_BACKUP_ZIP
su git -c "$GITEA_BIN dump --file $GITEA_BACKUP_ZIP"
rclone sync $GITEA_BACKUP_ZIP $RCLONE_TARGET