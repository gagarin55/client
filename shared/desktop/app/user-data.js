// @flow
import jsonfile from 'jsonfile'
import path from 'path'
import {app} from 'electron'

export default class UserData<State> {
  state: State
  file: string

  constructor(file: string, defaultState: State) {
    this.file = file
    this.state = defaultState
    try {
      let state = jsonfile.readFileSync(this.configPath())
      this.state = state
    } catch (err) {
      console.warn('Error loading state:', err)
    }
  }

  configPath = (): string => {
    return path.join(app.getPath('userData'), this.file)
  }

  save = () => {
    try {
      jsonfile.writeFile(this.configPath(), this.state)
    } catch (err) {
      console.warn('Error saving state:', err)
    }
  }
}
