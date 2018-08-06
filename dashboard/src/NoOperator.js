import { Container, Message, Modal, Segment } from 'semantic-ui-react';
import React from 'react';

const NoOperator = () => (
  <Container>
    <Modal open>
      <Modal.Header>Welcome to Kube-ArangoDB</Modal.Header>
      <Modal.Content>
        <Segment basic>
          <Message color="orange">
            There are no operators available yet.
          </Message>
        </Segment>
        {this.props.podInfoView}
        {(this.props.error) ? <Message error content={this.props.error}/> : null}
      </Modal.Content>
    </Modal>
  </Container>
);

export default NoOperator;
