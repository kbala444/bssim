#!/bin/sh
# this is probably a shitty way of doing this but idk sql
rm testing
rm testing2
rm testing-journal
rm testing2-journal

sqlite3 testing < ../schema.sql
sqlite3 testing2 < ../schema.sql
