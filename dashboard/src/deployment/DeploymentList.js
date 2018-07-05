import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import { Icon, Popup, Table } from 'semantic-ui-react';
import Loading from '../util/Loading.js';
import CommandInstruction from '../util/CommandInstruction.js';

const HeaderView = () => (
  <Table.Header>
    <Table.Row>
      <Table.HeaderCell>State</Table.HeaderCell>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Mode</Table.HeaderCell>
      <Table.HeaderCell>Version</Table.HeaderCell>
      <Table.HeaderCell><Popup trigger={<span>Pods</span>}>Ready / Total</Popup></Table.HeaderCell>
      <Table.HeaderCell><Popup trigger={<span>Volumes</span>}>Bound / Total</Popup></Table.HeaderCell>
      <Table.HeaderCell>StorageClass</Table.HeaderCell>
      <Table.HeaderCell></Table.HeaderCell>
    </Table.Row>
  </Table.Header>
);

const DatabaseLinkView = ({name, url}) => (
  <a href={url} target={name}>
    <Popup trigger={<Icon link name="database"/>}>
      Go the the web-UI of the database.
    </Popup>
  </a>
);

const NoDatabaseLinkView = () => (
  <Popup trigger={<Icon link name="database"/>}>
    This deployment is not reachable outside the Kubernetes cluster.
  </Popup>
);

const RowView = ({name, mode, version, license, ready_pod_count, pod_count, ready_volume_count, volume_count, storage_classes, database_url, delete_command}) => (
  <Table.Row>
    <Table.Cell><Icon name="bell" color="red"/></Table.Cell>
    <Table.Cell>{name}</Table.Cell>
    <Table.Cell>{mode}</Table.Cell>
    <Table.Cell>{version} {(license) ? `(${license})` : "" }</Table.Cell>
    <Table.Cell>{ready_pod_count} / {pod_count}</Table.Cell>
    <Table.Cell>{ready_volume_count} / {volume_count}</Table.Cell>
    <Table.Cell>{storage_classes.map((item) => (item === "") ? "<default>" : item)}</Table.Cell>
    <Table.Cell>
      { database_url ? <DatabaseLinkView name={name} url={database_url}/> : <NoDatabaseLinkView/>}
      <CommandInstruction 
        trigger={<Icon floated="right" name="trash alternate"/>}
        command={delete_command}
        title="Delete deployment"
        description="To delete this deployment, run:"
      />
    </Table.Cell>
  </Table.Row>
);

const ListView = ({items}) => (
  <Table striped celled>
    <HeaderView/>
    <Table.Body>
      {
        (items) ? items.map((item) => 
          <RowView 
            key={item.name} 
            name={item.name}
            namespace={item.namespace}
            mode={item.mode}
            version={item.database_version}
            license={item.database_license}
            ready_pod_count={item.ready_pod_count}
            pod_count={item.pod_count}
            ready_volume_count={item.ready_volume_count}
            volume_count={item.volume_count}
            storage_classes={item.storage_classes}
            database_url={item.database_url}
            delete_command={createDeleteCommand(item.name, item.namespace)}
          />) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

const EmptyView = () => (<div>No deployments</div>);

function createDeleteCommand(name, namespace) {
  return `kubectl delete ArangoDeployment -n ${namespace} ${name}`;
}

class DeploymentList extends Component {
  state = {};

  componentDidMount() {
    this.intervalId = setInterval(this.reloadDeployments, 5000);
    this.reloadDeployments();
  }

  componentWillUnmount() {
    clearInterval(this.intervalId);
  }

  reloadDeployments = async() => {
    const result = await apiGet('/api/deployment');
    this.setState({items:result.deployments});
  }

  render() {
    const items = this.state.items;
    if (!items) {
      return (<Loading/>);
    }
    if (items.length === 0) {
      return (<EmptyView/>);
    }
    return (<ListView items={items}/>);
  }
}

export default DeploymentList;
