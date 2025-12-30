# local development

## useful commands
- tree command , -I flag to ignore files/dirs
    - **FIRST, LOAD THIS ENV VAR:**
    - `
export EXCTR='*.sql|z_*|*.md|sql'
    `
        - `
tree -I $EXCTR
        `
    - save to file in z_dev dir:
        - `
tree -I $EXCTR > z_dev/tree.txt
        `
    - output to mac clipboard:
        - `
tree -I $EXCTR | pbcopy
        `