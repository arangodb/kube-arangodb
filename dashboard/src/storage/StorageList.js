import { Loader, Table } from 'semantic-ui-react';
import React, { Component } from 'react';
import ReactTimeout from 'react-timeout';

import { LoaderBoxForTable as LoaderBox } from '../style/style';
import { withAuth } from '../auth/Auth';
import api, { isUnauthorized } from '../api/api';
import Loading from '../util/Loading';
import StorageRow from './StorageRow';

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

const ListView = ({items, loading}) => (
  <Table celled>
    <HeaderView loading={loading}/>
    <Table.Body>
      {
        (items) ? items.map((item) => 
          <StorageRow 
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

const EmptyView = () => (<div>No local storage resources</div>);

function createDeleteCommand(name) {
  return `kubectl delete ArangoLocalStorage ${name}`;
}

function createDescribeCommand(name) {
  return `kubectl describe ArangoLocalStorage ${name}`;
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
      this.setState({
        loading: true
      });
      const result = await api.get('/api/storage');
      this.setState({
        items: result.storages,
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
    this.props.setTimeout(this.reloadStorages, 5000);
  }

  render() {
    const items = this.state.items;
    if (!items) {
      return (<Loading />);
    }
    if (items.length === 0) {
      return (<EmptyView />);
    }
    return (<ListView items={items} loading={this.state.loading} />);
  }
}

export default ReactTimeout(withAuth(StorageList));
