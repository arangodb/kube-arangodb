import React, { Component } from 'react';
import { apiGet } from '../api/api.js';
import { Icon, Table } from 'semantic-ui-react';
import Loading from '../util/Loading.js';

const HeaderView = () => (
  <Table.Header>
    <Table.Row>
      <Table.HeaderCell>State</Table.HeaderCell>
      <Table.HeaderCell>Name</Table.HeaderCell>
      <Table.HeaderCell>Mode</Table.HeaderCell>
      <Table.HeaderCell>Pods</Table.HeaderCell>
    </Table.Row>
  </Table.Header>
);

const RowView = ({name, mode, ready_pod_count, pod_count}) => (
  <Table.Row>
    <Table.Cell><Icon name="bell" color="red"/></Table.Cell>
    <Table.Cell>{name}</Table.Cell>
    <Table.Cell>{mode}</Table.Cell>
    <Table.Cell>{ready_pod_count} / {pod_count}</Table.Cell>
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
            mode={item.mode}
            ready_pod_count={item.ready_pod_count}
            pod_count={item.pod_count}
          />) : <p>No items</p>
      }
    </Table.Body>
  </Table>
);

const EmptyView = () => (<div>No deployments</div>);

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
