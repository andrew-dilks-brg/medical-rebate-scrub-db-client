### Example Usage
* `go run . -manu=Amgen -table=ndc_lu -step=add -file=/Users/Adilks/Downloads/NDC_LU2.csv`
* `go run . -manu=Amgen -table=ndc_lu -step=delete`
* `go run . -step=get -table=ndc_lu`
* `go run . -step=get -table=pos_lu -manu=amgen`


### These are the scripts used to create the tables themselves
* Added into schema "rbtbin"
    * Create NDC LU table
        "CREATE TABLE rbtbin.ndc_lu ( manu text, product text, ndc text, PRIMARY KEY ndc );"

    * Create POS_LU table
        "CREATE TABLE rbtbin.pos_lu ( manu text, pos text, PRIMARY KEY (manu, pos) );"

    * Create MOD_LU table
        "CREATE TABLE rbtbin.mod_lu ( manu text, mod_340b text, PRIMARY KEY (manu, mod_340b) );"

    * Create HCPCS_LU table
        "CREATE TABLE rbtbin.hcpcs_lu ( manu text, product text, hcpcs_cd text, PRIMARY KEY (manu, hcpcs_cd) );"

    * Create CSR_LIST table
        "CREATE TABLE rbtbin.csr_list ( manu text, npi text, product text, csr text, start_date text, term_date text, PRIMARY KEY (manu, npi, product) );"
