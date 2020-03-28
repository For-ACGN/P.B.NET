package msfrpc

// CoreModuleStats is used to return the number of modules loaded, broken down by type.
func (msf *MSFRPC) CoreModuleStats() (*CoreModuleStatsResult, error) {
	request := CoreModuleStatsRequest{
		Method: MethodCoreModuleStats,
		Token:  msf.GetToken(),
	}
	var result CoreModuleStatsResult
	err := msf.send(msf.ctx, &request, &result)
	if err != nil {
		return nil, err
	}
	if result.Err {
		return nil, &result.MSFError
	}
	return &result, nil
}

// CoreAddModulePath is used to add a new local file system directory (local to the server)
// as a module path. This can be used to dynamically load a separate module tree through
// the API. The path must be accessible to the user ID running the Metasploit service and
// contain a top-level directory for each module type (exploits, nop, encoder, payloads,
// auxiliary, post). Module paths will be immediately scanned for new modules and modules
// that loaded successfully will be immediately available. Note that this will not unload
// modules that were deleted from the file system since previously loaded (to remove all
// deleted modules, the core.reload_modules method should be used instead). This module
// may raise an error response if the specified path does not exist.
func (msf *MSFRPC) CoreAddModulePath(path string) (*CoreAddModulePathResult, error) {
	request := CoreAddModulePathRequest{
		Method: MethodCoreAddModulePath,
		Token:  msf.GetToken(),
		Path:   path,
	}
	var result CoreAddModulePathResult
	err := msf.send(msf.ctx, &request, &result)
	if err != nil {
		return nil, err
	}
	if result.Err {
		return nil, &result.MSFError
	}
	return &result, nil
}