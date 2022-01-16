#!/bin/sh
mdfile=$1;
if [ ${mdfile#*.} = 'md' ]; then
  style=$(grep -oP '<!--\s*[sS]tylesheet:\s*\K.*\.css\s*-->' $mdfile | cut -d' ' -f1);
  tmpmd=$(mktemp --suffix=.md /tmp/mdstyled-XXXXXXXXX);
  cat <(echo "<style>") $style <(echo "</style>") $mdfile > $tmpmd;
  echo $tmpmd && exit 0
fi
exit 1;
