`go build tf2_generate_items`

linux

`SET GOOS=linux&&SET GOARCH=amd64&&go build  tf2_generate_items`

# Usage
`tf2_generate_items -o <outputdir> -i <itemsdir> -r <resourcedir> -m=<0|1> -l <lang>`

itemsdir is usually <TF2 INSTALL DIR>/scripts/items/

resourcedir is usually <TF2 INSTALL DIR>/resource/

lang can be any supported tf2 language

-m to generate items or tournament medals
