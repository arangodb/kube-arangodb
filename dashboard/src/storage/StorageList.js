import { Accordion, Header, Icon, Loader, Popup, Table } from 'semantic-ui-react';
import api from '../api/api.js';
import CommandInstruction from '../util/CommandInstruction.js';
import VolumeList from './VolumeList.js';
import Loading from '../util/Loading.js';
import React, { Component } from 'react';
import ReactTimeout from 'react-timeout';
import styled from 'react-emotion';

const LoaderBox = styled('span')`
  float: right;
  width: 0;
  padding-right: 1em;
  max-width: 0;
  display: inline-block;
`;

const HeaderView = ({loading}) => (
  <Table.Header>
    <Table.Row>
      <Table.HeaderCell>State</Table.HeaderCell>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Local path(s)</Table.HeaderCell>
      <Table.HeaderCell>StorageClass</Table.HeaderCell>
      <Table.HeaderCell>
        Actions
        <LoaderBox><Loader size="mini" active={loading} inline/></LoaderBox>
      </Table.HeaderCell>
    </Table.Row>
  </Table.Header>
);

const RowView = ({name, stateColor,localPaths, storageClass, storageClassIsDefault, deleteCommand, describeCommand, expanded, toggleExpand}) => (
  <Table.Row>
    <Table.Cell>
      <Popup trigger={<Icon name={(stateColor==="green") ? "check" : "bell"} color={stateColor}/>}>
        {getStateColorDescription(stateColor)}
      </Popup>
    </Table.Cell>
    <Table.Cell onClick={toggleExpand}>
      <Accordion>
        <Accordion.Title active={expanded}>
          <Icon name='dropdown' /> 
          {name}
        </Accordion.Title>
      </Accordion>
    </Table.Cell>
    <Table.Cell>
      {localPaths.map((item) => <code>{item}</code>)}
    </Table.Cell>
    <Table.Cell>
      {storageClass}
      <span style={{"float":"right"}}>
        {storageClassIsDefault && <Popup trigger={<Icon name="exclamation"/>} content="Default storage class"/>}
      </span>
    </Table.Cell>
    <Table.Cell>
      <CommandInstruction 
          trigger={<Icon link name="zoom"/>}
          command={describeCommand}
          title="Describe local storage"
          description="To get more information on the state of this local storage, run:"
        />
      <span style={{"float":"right"}}>
        <CommandInstruction 
          trigger={<Icon link name="trash"/>}
          command={deleteCommand}
          title="Delete local storage"
          description="To delete this local storage, run:"
        />
      </span>
    </Table.Cell>
  </Table.Row>
);

const VolumesRowView = ({name}) => (
  <Table.Row>
    <Table.Cell colspan="5">
      <Header sub>Volumes</Header>
      <VolumeList storageName={name}/>
    </Table.Cell>
  </Table.Row>
);

const ListView = ({items, loading}) => (
  <Table celled>
    <HeaderView loading={loading}/>
    <Table.Body>
      {
        (items) ? items.map((item) => 
          <RowComponent 
            key={item.name} 
            name={item.name}
            localPaths={item.local_paths}
            stateColor={item.state_color}
            storageClass={item.storage_class}
            storageClassIsDefault={item.storage_class_is_default}
            deleteCommand={createDeleteCommand(item.name)}
            describeCommand={createDescribeCommand(item.name)}
          />
        ) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

class RowComponent extends Component {
  state = {expanded: true};

  onToggleExpand = () => { this.setState({expanded: !this.state.expanded});}

  render() {
    return [<RowView 
      key={this.props.name} 
      name={this.props.name}
      localPaths={this.props.localPaths}
      stateColor={this.props.stateColor}
      storageClass={this.props.storageClass}
      storageClassIsDefault={this.props.storageClassIsDefault}
      deleteCommand={this.props.deleteCommand}
      describeCommand={this.props.describeCommand}
      toggleExpand={this.onToggleExpand}
      expanded={this.state.expanded}
    />,
    this.state.expanded && <VolumesRowView
      key={`${this.props.name}-vol`} 
      name={this.props.name}
      expanded={this.state.expanded}
      toggleExpand={this.onToggleExpand}
    />
    ];
  }
}

const EmptyView = () => (<div>No local storage resources</div>);

function createDeleteCommand(name) {
  return `kubectl delete ArangoLocalStorage ${name}`;
}

function createDescribeCommand(name) {
  return `kubectl describe ArangoLocalStorage ${name}`;
}

function getStateColorDescription(stateColor) {
  switch (stateColor) {
    case "green":
      return "Everything is running smooth.";
    case "yellow":
      return "There is some activity going on, but local storage is available.";
    case "orange":
      return "There is some activity going on, local storage may be/become unavailable. You should pay attention now!";
    case "red":
      return "The local storage is in a bad state and manual intervention is likely needed.";
    default:
      return "State is not known.";
  }
}

class StorageList extends Component {
  state = {
    items: undefined,
    error: undefined,
    loading: true
  };

  componentDidMount() {
    this.reloadStorages();
  }

  reloadStorages = async() => {
    try {
      this.setState({loading: true});
      const result = await api.get('/api/storage');
      this.setState({
        items: result.storages,
        loading: false,
        error: undefined
      });
    } catch (e) {
      this.setState({error: e.message, loading: false});
    }
    this.props.setTimeout(this.reloadStorages, 5000);
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

export default ReactTimeout(StorageList);
