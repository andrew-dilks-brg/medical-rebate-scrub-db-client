### Overview
* This is a tool that was created to interact with the rbtbin - MRS tables. It has read, create, and delete apis for 5 tables that MRS relies on. Can be used by anyone with the correct SDM permissions to update our source of truth for MRS scrubs in the rebate-binary.

### Example Usage
* `go run . -manu=amgen -table=mrs_ndc_lu -step=add -file=/Users/Adilks/Downloads/NDC_LU2.csv`
* `go run . -manu=amgen -table=mrs_ndc_lu -step=add -file=./NDC_LU.csv`
* `go run . -manu=amgen -table=mrs_ndc_lu -step=delete`
* `go run . -step=get -table=mrs_ndc_lu`
* `go run . -step=get -table=pos_lu -manu=amgen`


### TODO
* Add bulk upload feature to improve upload time for CSR and CBKS datasets - these take 10min with single writes per row

### These are the scripts used to create the tables themselves
* Added into schema "rbtbin"
    * Create NDC LU table
        "CREATE TABLE rbtbin.mrs_ndc_lu ( manufacturer text, product text, ndc text, PRIMARY KEY (ndc) );"

    * Create POS_LU table
        "CREATE TABLE rbtbin.mrs_pos_lu ( manufacturer text, pos text, pos_type text, PRIMARY KEY (manufacturer, pos) );"

    * Create MOD_LU table
        "CREATE TABLE rbtbin.mrs_mod_lu ( manufacturer text, mod_340b text, mod_type text, PRIMARY KEY (manufacturer, mod_340b) );"

    * Create HCPCS_LU table
        "CREATE TABLE rbtbin.mrs_hcpcs_lu ( manufacturer text, product text, hcpcs_cd text, PRIMARY KEY (manufacturer, hcpcs_cd) );"

    * Create CSR_LIST table
        "CREATE TABLE rbtbin.mrs_csr_list ( manufacturer text, npi text, product text, csr text, start_date text, term_date text, PRIMARY KEY (manufacturer, npi, product) );"

    * Create BINARY_CHARGEBACKS table
        "CREATE TABLE rbtbin.mrs_binary_cbks ( manufacturer text, product text, npi text, description text, priority text, PRIMARY KEY (manufacturer, product, npi) );"
