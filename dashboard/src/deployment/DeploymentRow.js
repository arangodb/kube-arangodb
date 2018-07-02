import React, { Component } from 'react';
import { Icon, Table } from 'semantic-ui-react';
//import logo from './logo.svg';
//import './App.css';

class DeploymentRow extends Component {
  render() {
    return (
      <Table.Row>
        <Table.Cell><Icon name="bell" color="red"/></Table.Cell>
        <Table.Cell>{this.props.info.name}</Table.Cell>
        <Table.Cell>{this.props.info.mode}</Table.Cell>
        <Table.Cell>{this.props.info.ready_pod_count} / {this.props.info.pod_count}</Table.Cell>
      </Table.Row>
    );
  }
}

export default DeploymentRow;
