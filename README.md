### Overview
* This is a tool that was created to interact with the rbtbin - MRS tables. It has read, create, and delete apis for 5 tables that MRS relies on. Can be used by anyone with the correct SDM permissions to update our source of truth for MRS scrubs in the rebate-binary.

### Example Usage
* `go run . -manu=amgen -table=ndc_lu -step=add -file=/Users/Adilks/Downloads/NDC_LU2.csv`
* `go run . -manu=amgen -table=ndc_lu -step=add -file=./NDC_LU.csv`
* `go run . -manu=amgen -table=ndc_lu -step=delete`
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
