dialects=("postgres" "mysql" "mssql" "sqlite" "hdb")

for dialect in "${dialects[@]}" ; do
    DEBUG=false GORM_DIALECT=${dialect} go test
done
