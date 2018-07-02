import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import DeploymentRow from './DeploymentRow.js';
import { Table } from 'semantic-ui-react';
//import logo from './logo.svg';
//import './App.css';

class DeploymentList extends Component {
  constructor() {
    super();
    this.state = {items:[]};
  }

  async componentDidMount() {
    this.reloadDeployments();
  }

  async reloadDeployments() {
    const result = await apiGet('/api/deployment');
    this.setState({items:result.deployments});
  }

  render() {
    setTimeout(this.reloadDeployments.bind(this), 5000);
    const items = this.state.items;
    if (items.length === 0) {
      return (<div>No deployments</div>);
    }
    return (
      <Table striped celled>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell>State</Table.HeaderCell>
            <Table.HeaderCell>Name</Table.HeaderCell>
            <Table.HeaderCell>Mode</Table.HeaderCell>
            <Table.HeaderCell>Pods</Table.HeaderCell>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {
            (items) ? items.map((item) => <DeploymentRow key={item.name} info={item}/>) : <p>No items</p>
          }
        </Table.Body>
      </Table>
    );
  }
}

export default DeploymentList;
