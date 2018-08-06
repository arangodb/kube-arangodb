import { Icon, Loader, Popup, Table } from 'semantic-ui-react';
import { Link } from "react-router-dom";
import React, { Component } from 'react';
import ReactTimeout from 'react-timeout';

import { LoaderBoxForTable as LoaderBox } from '../style/style';
import { withAuth } from '../auth/Auth';
import api, { isUnauthorized } from '../api/api';
import CommandInstruction from '../util/CommandInstruction';
import Loading from '../util/Loading';

const HeaderView = ({loading}) => (
  <Table.Header>
    <Table.Row>
      <Table.HeaderCell>State</Table.HeaderCell>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Mode</Table.HeaderCell>
      <Table.HeaderCell>Version</Table.HeaderCell>
      <Table.HeaderCell><Popup trigger={<span>Pods</span>}>Ready / Total</Popup></Table.HeaderCell>
      <Table.HeaderCell><Popup trigger={<span>Volumes</span>}>Bound / Total</Popup></Table.HeaderCell>
      <Table.HeaderCell>StorageClass</Table.HeaderCell>
      <Table.HeaderCell>
        Actions
        <LoaderBox><Loader size="mini" active={loading} inline/></LoaderBox>
      </Table.HeaderCell>
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

const RowView = ({name, mode, environment, stateColor, version, license, readyPodCount, podCount, readyVolumeCount, volumeCount, storageClasses, databaseURL, deleteCommand, describeCommand}) => (
  <Table.Row>
    <Table.Cell>
      <Popup trigger={<Icon name={(stateColor==="green") ? "check" : "bell"} color={stateColor}/>}>
        {getStateColorDescription(stateColor)}
      </Popup>
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
    <Table.Cell>{readyPodCount} / {podCount}</Table.Cell>
    <Table.Cell>{readyVolumeCount} / {volumeCount}</Table.Cell>
    <Table.Cell>{storageClasses.map((item) => (item === "") ? "<default>" : item)}</Table.Cell>
    <Table.Cell>
      { databaseURL ? <DatabaseLinkView name={name} url={databaseURL}/> : <NoDatabaseLinkView/>}
      <CommandInstruction 
          trigger={<Icon link name="zoom"/>}
          command={describeCommand}
          title="Describe deployment"
          description="To get more information on the state of this deployment, run:"
        />
      <span style={{"float":"right"}}>
        <CommandInstruction 
          trigger={<Icon link name="trash"/>}
          command={deleteCommand}
          title="Delete deployment"
          description="To delete this deployment, run:"
        />
      </span>
    </Table.Cell>
  </Table.Row>
);

const ListView = ({items, loading}) => (
  <Table striped celled>
    <HeaderView loading={loading}/>
    <Table.Body>
      {
        (items) ? items.map((item) => 
          <RowView 
            key={item.name} 
            name={item.name}
            namespace={item.namespace}
            mode={item.mode}
            environment={item.environment}
            stateColor={item.state_color}
            version={item.database_version}
            license={item.database_license}
            readyPodCount={item.ready_pod_count}
            podCount={item.pod_count}
            readyVolumeCount={item.ready_volume_count}
            volumeCount={item.volume_count}
            storageClasses={item.storage_classes}
            databaseURL={item.database_url}
            deleteCommand={createDeleteCommand(item.name, item.namespace)}
            describeCommand={createDescribeCommand(item.name, item.namespace)}
          />) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

const EmptyView = () => (<div>No deployments</div>);

function createDeleteCommand(name, namespace) {
  return `kubectl delete ArangoDeployment -n ${namespace} ${name}`;
}

function createDescribeCommand(name, namespace) {
  return `kubectl describe ArangoDeployment -n ${namespace} ${name}`;
}

function getStateColorDescription(stateColor) {
  switch (stateColor) {
    case "green":
      return "Everything is running smooth.";
    case "yellow":
      return "There is some activity going on, but deployment is available.";
    case "orange":
      return "There is some activity going on, deployment may be/become unavailable. You should pay attention now!";
    case "red":
      return "The deployment is in a bad state and manual intervention is likely needed.";
    default:
      return "State is not known.";
  }
}

class DeploymentList extends Component {
  state = {
    items: null,
    error: null,
    loading: true
  };

  componentDidMount() {
    this.reloadDeployments();
  }

  reloadDeployments = async() => {
    try {
      this.setState({loading: true});
      const result = await api.get('/api/deployment');
      this.setState({
        items: result.deployments,
        loading: false,
        error: null
      });
    } catch (e) {
      this.setState({error: e.message, loading: false});
      if (isUnauthorized(e)) {
        this.props.doLogout();
        return;
      }
    }
    this.props.setTimeout(this.reloadDeployments, 5000);
  }

  render() {
    const items = this.state.items;
    if (!items) {
      return (<Loading/>);
    }
    if (items.length === 0) {
      return (<EmptyView/>);
    }
    return (<ListView items={items} loading={this.state.loading}/>);
  }
}

export default ReactTimeout(withAuth(DeploymentList));
