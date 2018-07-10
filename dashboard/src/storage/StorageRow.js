import { Accordion, Header, Icon, Popup, Table } from 'semantic-ui-react';
import React, { Component } from 'react';

import CommandInstruction from '../util/CommandInstruction';
import VolumeList from './VolumeList';

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
      {localPaths.map((item, index) => <code key={index}>{item}</code>)}
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
    <Table.Cell colSpan="5">
      <Header sub>Volumes</Header>
      <VolumeList storageName={name}/>
    </Table.Cell>
  </Table.Row>
);

class StorageRow extends Component {
  state = {expanded: true};

  onToggleExpand = () => { this.setState({expanded: !this.state.expanded});}

  render() {
    return [<RowView 
      key="datarow"
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
      key="volrow"
      name={this.props.name}
      expanded={this.state.expanded}
      toggleExpand={this.onToggleExpand}
    />
    ];
  }
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

export default StorageRow;
