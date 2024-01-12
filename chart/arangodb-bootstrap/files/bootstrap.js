//
// DISCLAIMER
//
// Copyright 2024 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//

const users = require('@arangodb/users');


function createUser(user, password) {
    if (createUserCond(user, password)) {
        console.log("User %s created", user)
    } else {
        console.log("User %s already exists, skip", user)
    }
}

function createUserCond(user, password) {
    try {
        users.save(user, password);
        return true
    } catch (e) {
        if (e.code === 409 && e.errorNum === 1702) {
            return false
        }
        throw e
    }

}

function grantCollection(database, collection, user, privilege) {
    if (grantCollectionCond(database, collection, user, privilege)) {
        console.log("Database %s/%s granted `%s` for %s", database, collection, privilege, user)
    } else {
        console.log("Database %s/%s grant for %s already exists, skip", database, collection, user)
    }
}

function grantCollectionCond(database, collection, user, privilege) {
    try {
        users.grantCollection(user, database, collection, privilege)
        return true
    } catch (e) {
        if (e.code === 409 && e.errorNum === 1207) {
            return false
        }
        throw e
    }
}

function grantDatabase(database, user, privilege) {
    if (grantDatabaseCond(database, user, privilege)) {
        console.log("Database %s granted `%s` for %s", database, privilege, user)
    } else {
        console.log("Database %s grant for %s already exists, skip", database, user)
    }
}

function grantDatabaseCond(database, user, privilege) {
    try {
        users.grantDatabase(user, database, privilege)
        return true
    } catch (e) {
        if (e.code === 409 && e.errorNum === 1207) {
            return false
        }
        throw e
    }
}

function createDatabase(database, options) {
    if (createDatabaseCond(database, options)) {
        console.log("Database %s created", database)
    } else {
        console.log("Database %s already exists, skip", database)
    }
}

function createDatabaseCond(database, options) {
    db._useDatabase("_system");
    try {
        db._createDatabase(database, options);
        return true
    } catch (e) {
        if (e.code === 409 && e.errorNum === 1207) {
            return false
        }
        throw e
    }
}

function createCollection(database, collection, properties, type) {
    if (createCollectionCond(database, collection, properties, type)) {
        console.log("Collection %s/%s created", database, collection)
    } else {
        console.log("Collection %s/%s already exists, skip", database, collection)
    }
}

function createCollectionCond(database, collection, properties, type) {
    db._useDatabase(database);
    try {
        db._create(collection, properties, type, {
            "waitForSyncReplication": true,
            "enforceReplicationFactor": true
        });
        return true
    } catch (e) {
        if (e.code === 409 && e.errorNum === 1207) {
            return false
        }
        throw e
    }
}

console.log("Starting ArangoDB Bootstrap")

db._version();
console.log("ArangoDB reachable");

console.log("Create Users")
{{- range $k, $v := .Values.users }}
createUser({{ $k | quote }}, process.env.PASSWORD_{{ $k | sha256sum | upper }});
{{- end }}

console.log("Create Databases")
{{- range $k, $v := .Values.databases }}
{{- if ne $k "*" }}
createDatabase({{ $k  | quote }}, {{ $v.options | default dict | toJson }});
{{- end }}
{{- range $u, $p := ($v.grants | default dict) }}
grantDatabase({{ $k | quote }}, {{ $u | quote }}, {{ $p | quote }});
{{- end }}
{{- range $c, $co := ($v.collections | default dict) }}
{{- if and (ne $k "*") (ne $c "*") }}
createCollection({{ $k | quote }}, {{ $c | quote }}, {{ $co.attributes | default dict | toJson }}, {{ $co.type | default "document" | quote }});
{{- end }}
{{- range $u, $p := ($co.grants | default dict) }}
grantCollection({{ $k | quote }}, {{ $c | quote }}, {{ $u | quote }}, {{ $p | quote }});
{{- end }}
{{- end }}
{{- end }}

console.log("Bootstrap completed")