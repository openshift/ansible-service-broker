#!/bin/env bash

create_gist () {
    local date=$1
    local notes_file=$2

	# 1. Somehow sanitize the file content
	#    Remove \r (from Windows end-of-lines),
	#    Replace tabs by \t
	#    Replace " by \"
	#    Replace EOL by \n
	CONTENT=$(sed -e 's/\r//' -e's/\t/\\t/g' -e 's/"/\\"/g' "${notes_file}" | awk '{ printf($0 "\\n") }')

	# 2. Build the JSON request
	read -r -d '' DESC <<EOF
	{
	  "description": "Release notes generated on $1",
	  "public": true,
	  "files": {
		"Release_Notes.md": {
		  "content": "${CONTENT}"
		}
	  }
	}
EOF

    $GH_CURL -X POST -d "${DESC}" $GIST_API
}

generate_notes () {
    local start=$1
    local end=HEAD
    local commits=$(git --no-pager log --reverse --pretty=format:"%s" "$start".."$end")
    local first_commit=$(git rev-parse "$start")
    local last_commit=$(git rev-parse "$end")
    local date=$(date +%Y%m%d)
    local notes_file=$(mktemp)
    local bug_file=$(mktemp)
    local other_file=$(mktemp)

    cat <<EOF >> $notes_file
# Release notes for $date

* First Commit: $first_commit
* Last Commit: $last_commit

EOF

    cat <<EOF >> $bug_file

## Bugs

EOF

    cat <<EOF >> $other_file

## Other Enhancements

EOF

    while read -r commit; do
        pr_number=$(echo $commit | sed -n 's/.*(#\([[:digit:]]\+\))$/\1/p')
        bug_number=$(echo $commit| sed -n 's/Bug\s\+\([[:digit:]]\+\).*/\1/p')
        if [ -n "$pr_number" ]; then
            detailed_commit=$(echo $commit | sed -n 's,\(.*\)(#\([[:digit:]]\+\))$,\1[(#\2)](https://github.com/openshift/ansible-service-broker/pull/\2),p')
            if [ ! -z "$bug_number" ]; then
                echo "* $detailed_commit" | sed -n 's,Bug\s\+\([[:digit:]]\+\)\(.*\),[Bug \1](https://bugzilla.redhat.com/show_bug.cgi?id=\1)\2,p' >> $bug_file
            else
                echo "* $detailed_commit" >> $other_file
            fi
        fi
    done <<< "$commits"

    cat $bug_file >> $notes_file
    rm $bug_file
    cat $other_file >> $notes_file
    rm $other_file

    create_gist $date $notes_file
    rm $notes_file
}

##############################
# Main
##############################
if [ ! -n "$GH_USER" ] || [ ! -n "$GH_TOKEN" ]; then
    echo "You must export GH_USER - your github user id"
    echo "You must export GH_TOKEN - your github access token"
    exit 1
fi

if [ ! -n "$1" ]; then
    echo "Must provide start tag/branch/sha to compare against"
    exit 1
fi

GH_CURL="curl -u $GH_USER:$GH_TOKEN"
GIST_API="https://api.github.com/gists"
generate_notes $1


###
# Example using github API
###
#ASB_GITHUB_API="https://api.github.com/repos/openshift/ansible-service-broker"
#STATE="closed"
#BRANCH="master"
#last_page=$($GH_CURL -sI $ASB_GITHUB_API/pulls\?state\=$STATE\&base\=$BRANCH | sed -nr 's/^Link:.*page=([0-9]+)>; rel="last".*/\1/p')
#echo "$last_page pages"
#for (( i=1; i<=$last_page; i++)); do
#    $GH_CURL -s $ASB_GITHUB_API/pulls\?state\=$STATE\&base\=$BRANCH\&page\=$i | \
#        jq '.[] | {
#            user: .user.login,
#            title: .title,
#            url: .html_url,
#            merged: .merged_at
#        }'
#done
