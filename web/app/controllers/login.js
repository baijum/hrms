import Ember from 'ember';

export default Ember.Controller.extend({
  session: Ember.inject.service('session'),

  actions: {
    authenticate() {
      var credentials = this.getProperties('identification', 'password');
      console.log(credentials);
      let { identification, password } = this.getProperties('identification', 'password');
      console.log(identification, password);
      this.get('session').authenticate('authenticator:jwt', identification, password).catch((reason) => {
        this.set('errorMessage', reason.error || reason);
      });
    }
  }

});
