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
      <Table.HeaderCell>Source</Table.HeaderCell>
      <Table.HeaderCell>Destination</Table.HeaderCell>
      <Table.HeaderCell>
        Actions
        <LoaderBox><Loader size="mini" active={loading} inline/></LoaderBox>
      </Table.HeaderCell>
    </Table.Row>
  </Table.Header>
);

const RowView = ({name, mode, stateColor, source, destination, deleteCommand, describeCommand}) => (
  <Table.Row>
    <Table.Cell>
      <Popup trigger={<Icon name={(stateColor==="green") ? "check" : "bell"} color={stateColor}/>}>
        {getStateColorDescription(stateColor)}
      </Popup>
    </Table.Cell>
    <Table.Cell>
      <Link to={`/deployment-replication/${name}`}>
        {name}
      </Link>
    </Table.Cell>
    <Table.Cell>
      {source}
    </Table.Cell>
    <Table.Cell>
      {destination}
    </Table.Cell>
    <Table.Cell>
      <CommandInstruction 
          trigger={<Icon link name="zoom"/>}
          command={describeCommand}
          title="Describe deployment replication"
          description="To get more information on the state of this deployment replication, run:"
        />
      <span style={{"float":"right"}}>
        <CommandInstruction 
          trigger={<Icon link name="trash"/>}
          command={deleteCommand}
          title="Delete deployment replication"
          description="To delete this deployment replication, run:"
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
            stateColor={item.state_color}
            source={item.source.deployment_name || item.source.master_endpoint}
            destination={item.destination.deployment_name || item.destination.master_endpoint}
            deleteCommand={createDeleteCommand(item.name, item.namespace)}
            describeCommand={createDescribeCommand(item.name, item.namespace)}
          />) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

const EmptyView = () => (<div>No deployment replications</div>);

function createDeleteCommand(name, namespace) {
  return `kubectl delete ArangoDeploymentReplication -n ${namespace} ${name}`;
}

function createDescribeCommand(name, namespace) {
  return `kubectl describe ArangoDeploymentReplication -n ${namespace} ${name}`;
}

function getStateColorDescription(stateColor) {
  switch (stateColor) {
    case "green":
      return "Replication has been configured.";
    case "yellow":
      return "Replication is being configured.";
    case "red":
      return "The replication is in a bad state and manual intervention is likely needed.";
    default:
      return "State is not known.";
  }
}

class DeploymentReplicationList extends Component {
  state = {
    items: null,
    error: null,
    loading: true
  };

  componentDidMount() {
    this.reloadDeploymentReplications();
  }

  reloadDeploymentReplications = async() => {
    try {
      this.setState({loading: true});
      const result = await api.get('/api/deployment-replication');
      this.setState({
        items: result.replications,
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
    this.props.setTimeout(this.reloadDeploymentReplications, 5000);
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

export default ReactTimeout(withAuth(DeploymentReplicationList));
