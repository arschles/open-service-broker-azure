package mysqldb

import (
	"fmt"

	"github.com/Azure/open-service-broker-azure/pkg/generate"
	"github.com/Azure/open-service-broker-azure/pkg/service"
)

func (s *serviceManager) ValidateBindingParameters(
	bindingParameters service.BindingParameters,
) error {
	// There are no parameters for binding to MySQL, so there is nothing
	// to validate
	return nil
}

func (s *serviceManager) Bind(
	_ service.StandardProvisioningContext,
	provisioningContext service.ProvisioningContext,
	bindingParameters service.BindingParameters,
) (service.BindingContext, service.Credentials, error) {
	pc, ok := provisioningContext.(*mysqlProvisioningContext)
	if !ok {
		return nil, nil, fmt.Errorf(
			"error casting provisioningContext as *mysqlProvisioningContext",
		)
	}

	userName := generate.NewIdentifier()
	password := generate.NewPassword()

	db, err := getDBConnection(pc)
	if err != nil {
		return nil, nil, err
	}
	defer db.Close() // nolint: errcheck

	// Open doesn't open a connection. Validate DSN data:
	if err = db.Ping(); err != nil {
		return nil, nil, err
	}

	if _, err = db.Exec(
		fmt.Sprintf("CREATE USER '%s'@'%%' IDENTIFIED BY '%s'", userName, password),
	); err != nil {
		return nil, nil, fmt.Errorf(
			`error creating user "%s": %s`,
			userName,
			err,
		)
	}

	if _, err = db.Exec(
		fmt.Sprintf("GRANT SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, RELOAD, "+
			"PROCESS, INDEX, ALTER, SHOW DATABASES, CREATE TEMPORARY TABLES, "+
			"LOCK TABLES, CREATE VIEW, SHOW VIEW, CREATE ROUTINE, ALTER ROUTINE, "+
			"CREATE USER, REFERENCES, EVENT, "+
			"TRIGGER ON *.* TO '%s'@'%%' WITH GRANT OPTION",
			userName)); err != nil {
		return nil, nil, fmt.Errorf(
			`error granting permission to "%s": %s`,
			userName,
			err,
		)
	}

	return &mysqlBindingContext{
			LoginName: userName,
		},
		&Credentials{
			Host:     pc.FullyQualifiedDomainName,
			Port:     3306,
			Database: pc.DatabaseName,
			Username: fmt.Sprintf("%s@%s", userName, pc.ServerName),
			Password: password,
		},
		nil
}
