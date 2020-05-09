Pod::Spec.new do |spec|
  spec.name         = 'Gcore'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/core-coin/go-core'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Core Client'
  spec.source       = { :git => 'https://github.com/core-coin/go-core.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gcore.framework'

	spec.prepare_command = <<-CMD
    curl https://gcorestore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gcore.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
