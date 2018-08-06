import { Icon, Loader, Popup, Table } from 'semantic-ui-react';
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
      <Table.HeaderCell>Capacity</Table.HeaderCell>
      <Table.HeaderCell>Node</Table.HeaderCell>
      <Table.HeaderCell>
        Actions
        <LoaderBox><Loader size="mini" active={loading} inline/></LoaderBox>
      </Table.HeaderCell>
    </Table.Row>
  </Table.Header>
);

const RowView = ({name, stateColor, nodeName, capacity, describeCommand, deleteCommand}) => (
  <Table.Row>
    <Table.Cell>
      <Popup trigger={<Icon name={(stateColor==="green") ? "check" : "bell"} color={stateColor}/>}>
        {getStateColorDescription(stateColor)}
      </Popup>
    </Table.Cell>
    <Table.Cell>
      {name}
    </Table.Cell>
    <Table.Cell>
      {capacity}
    </Table.Cell>
    <Table.Cell>
      {nodeName}
    </Table.Cell>
    <Table.Cell>
      <CommandInstruction 
          trigger={<Icon link name="zoom"/>}
          command={describeCommand}
          title="Describe PersistentVolume"
          description="To get more information on the state of this PersistentVolume, run:"
        />
      <span style={{"float":"right"}}>
        <CommandInstruction 
          trigger={<Icon link name="trash"/>}
          command={deleteCommand}
          title="Delete PersistentVolume"
          description="To delete this PersistentVolume, run:"
        />
      </span>
    </Table.Cell>
  </Table.Row>
);

const ListView = ({items, loading}) => (
  <Table celled>
    <HeaderView loading={loading}/>
    <Table.Body>
      {
        (items) ? items.map((item) => 
          <RowView 
            key={item.name} 
            name={item.name}
            stateColor={item.state_color}
            nodeName={item.node_name}
            capacity={item.capacity}
            deleteCommand={createDeleteCommand(item.name)}
            describeCommand={createDescribeCommand(item.name)}
          />
        ) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

const EmptyView = () => (<div>No PersistentVolumes</div>);

function createDeleteCommand(name) {
  return `kubectl delete PersistentVolume ${name}`;
}

function createDescribeCommand(name) {
  return `kubectl describe PersistentVolume ${name}`;
}

function getStateColorDescription(stateColor) {
  switch (stateColor) {
    case "green":
      return "Everything is running smooth.";
    case "yellow":
      return "There is some activity going on, but PersistentVolume is available.";
    case "orange":
      return "There is some activity going on, PersistentVolume may be/become unavailable. You should pay attention now!";
    case "red":
      return "The PersistentVolume is in a bad state and manual intervention is likely needed.";
    default:
      return "State is not known.";
  }
}

class VolumeList extends Component {
  state = {
    items: undefined,
    error: undefined,
    loading: true
  };

  componentDidMount() {
    this.reloadVolumes();
  }

  reloadVolumes = async() => {
    try {
      this.setState({
        loading: true
      });
      const result = await api.get(`/api/storage/${this.props.storageName}`);
      this.setState({
        items: result.volumes,
        loading: false,
        error: undefined
      });
    } catch (e) {
      this.setState({
        error: e.message,
        loading: false
      });
      if (isUnauthorized(e)) {
        this.props.doLogout();
        return;
      }
    }
    this.props.setTimeout(this.reloadVolumes, 5000);
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

export default ReactTimeout(withAuth(VolumeList));
