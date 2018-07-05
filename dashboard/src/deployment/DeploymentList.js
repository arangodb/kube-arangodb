import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import { Icon, Popup, Table } from 'semantic-ui-react';
import Loading from '../util/Loading.js';
import CommandInstruction from '../util/CommandInstruction.js';
import { Link } from "react-router-dom";

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

const RowView = ({name, mode, environment, version, license, ready_pod_count, pod_count, ready_volume_count, volume_count, storage_classes, database_url, delete_command}) => (
  <Table.Row>
    <Table.Cell>
      <Icon name="bell" color="red"/>
    </Table.Cell>
    <Table.Cell>
      <Link to={`/deployment/${name}`}>
        {name}
      </Link>
    </Table.Cell>
    <Table.Cell>
      {mode}
      <span style={{"float":"right"}}>
        {(environment==="Development") ? <Popup trigger={<Icon name="laptop"/>} content="Development environment"/>: null}
        {(environment==="Production") ? <Popup trigger={<Icon name="warehouse"/>} content="Production environment"/>: null}
      </span>
    </Table.Cell>
    <Table.Cell>
      {version}
      <span style={{"float":"right"}}>
        {(license==="community") ? <Popup trigger={<Icon name="users"/>} content="Community edition"/>: null}
        {(license==="enterprise") ? <Popup trigger={<Icon name="dollar"/>} content="Enterprise edition"/>: null}
      </span>
    </Table.Cell>
    <Table.Cell>{ready_pod_count} / {pod_count}</Table.Cell>
    <Table.Cell>{ready_volume_count} / {volume_count}</Table.Cell>
    <Table.Cell>{storage_classes.map((item) => (item === "") ? "<default>" : item)}</Table.Cell>
    <Table.Cell>
      { database_url ? <DatabaseLinkView name={name} url={database_url}/> : <NoDatabaseLinkView/>}
      <span style={{"float":"right"}}>
        <CommandInstruction 
          trigger={<Icon link name="trash"/>}
          command={delete_command}
          title="Delete deployment"
          description="To delete this deployment, run:"
        />
      </span>
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
            environment={item.environment}
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
